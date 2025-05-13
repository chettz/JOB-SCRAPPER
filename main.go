package main

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

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=python"

// make a goroutines using channels to make fast
// main <-> getPage, getPage <-> extractJob

func main() {
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages()

	for i:=0;i<totalPages;i++{
		go getPage(i+1, c)
	}

	for i:=0;i<totalPages;i++{
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := baseURL + "&recruitPage=" + strconv.Itoa(page)
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
	title := cleanString(card.Find(".job_tit>a").Text())
	card.Find(".job_condition span").First().Find("a").Each(func(i int, s *goquery.Selection){
		locations = append(locations, s.Text())
		location = strings.Join(locations, " ") // ["seoul", "yongsan-gu"] -> "seoul yongsan-gu"
	})
	corp := cleanString(card.Find(".corp_name").Text())
	date := cleanString(card.Find(".date").Text())
	c <- extractedJob{
		id: 		id, 
		title: 		title, 
		location: 	location, 
		corp: 		corp, 
		date: 		date}
}

func cleanString(str string) string {
	// removing space from the both side and seperating all the words removes the space between text
	// the texts parsed from html contains spaces between words. it can be removed by strings.Fields and made to array 
	// Strings.Join(array -> string) => concatenating the elements of a to create a single string with a seperater
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
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
	file, err := os.Create("jobs.csv")
	checkErr((err))

	w := csv.NewWriter(file) // buffer
	defer w.Flush() // run when functions ends

	headers := []string{"ID", "Title", "Date", "Location", "Corp"}

	wErr := w.Write(headers)
	checkErr(wErr)
	
	for _, job := range jobs{
		links := "https://www.saramin.co.kr/zf_user/jobs/relay/view?isMypage=no&rec_idx=" + job.id
		jobSlice := []string{links, job.title, job.date, job.location, job.corp}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

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