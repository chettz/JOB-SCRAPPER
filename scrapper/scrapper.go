package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct{
	id			string
	title		string
	date		string  //deadline
	location 	string
	corp 		string
	
}


// make a goroutines using channels to make fast
// main <-> getPage(goroutines * 10(totalpages)), getPage <-> extractJob (goroutines * # of jobs per page ) | total goroutines 10 + (10*40) 
// make writeJobs to goroutines

// Scrape saramin by a term
func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=" + term
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)

	for i:=0;i<totalPages;i++{
		go getPage(i+1, baseURL, c)
	}

	for i:=0; i<totalPages; i++{
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")

	searchCards.Each(func(i int, card *goquery.Selection){ // 'card' means each card section
		go extractJob(card, c) // calls multiple extractJob at the same time
	})

	for i:=0; i<searchCards.Length(); i++{
		job := <-c // recieve the results, block operation
		jobs = append(jobs, job)
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	locations := []string{}
	var location string
	id, _ := card.Attr("value")
	title := CleanString(card.Find(".job_tit>a").Text())
	card.Find(".job_condition span").First().Find("a").Each(func(i int, s *goquery.Selection){
		locations = append(locations, s.Text())
		location = strings.Join(locations, " ") // ["seoul", "yongsan-gu"] -> "seoul yongsan-gu"
	})
	corp := CleanString(card.Find(".corp_name").Text())
	date := CleanString(card.Find(".date").Text())
	c <- extractedJob{
		id: 		id, 
		title: 		title, 
		location: 	location, 
		corp: 		corp, 
		date: 		date}
}

// CleanString cleans a string
func CleanString(str string) string {
	// removing space from the both side and seperating all the words removes the space between text
	// the texts parsed from html contains spaces between words. it can be removed by strings.Fields and made to array 
	// Strings.Join(array -> string) => concatenating the elements of a to create a single string with a seperater
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})
	return pages
}

func writeJobs(jobs []extractedJob) {
	c := make(chan []string)
	file, err := os.Create("jobs.csv")
	checkErr((err))
	// prevent the corruption of hangul
	utf8bom := []byte{0xEF, 0xBB, 0xBF} 
	file.Write(utf8bom) // file.Write() is appropriate at writing little amount of data. It's not effective at repetitive writing causing disk I/O. 

	w := csv.NewWriter(file) // buffer
	defer w.Flush() // run when functions ends, writes all contents in the buffer at once.
	defer file.Close() // clear FD

	headers := []string{"ID", "Title", "Date", "Location", "Corp"}

	wErr := w.Write(headers)
	checkErr(wErr)
	for _, job := range jobs{
		go writingJobs(job, c)
	}

	for i:=0; i<len(jobs);i++{
		w.Write(<-c)
	}
}

func writingJobs(job extractedJob, c chan<- []string) {
		links := "https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.id
		jobSlice := []string{links, job.title, job.date, job.location, job.corp}
		c <- jobSlice
}

func checkErr(err error){
	if err != nil{
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response){
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}