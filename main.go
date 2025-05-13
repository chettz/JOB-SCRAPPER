package main

import (
	"fmt"
	"log"
	"net/http"
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


func main() {
	var jobs []extractedJob
	totalPages := getPages()

	for i:=0;i<totalPages;i++{
		extractedJobs := getPage(i+1)
		jobs = append(jobs, extractedJobs...) // get the contents
	}

	fmt.Println(jobs)
}

func getPage(page int) []extractedJob {
	var jobs []extractedJob
	pageURL := baseURL + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// searchCards := doc.Find(".item_recruit")

	doc.Find(".item_recruit").Each(func(i int, card *goquery.Selection){ // 's' means each card section
		job := extractJob(card)
		jobs = append(jobs, job)
	})
	return jobs
}

func extractJob(card *goquery.Selection) extractedJob {
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
	return extractedJob{
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