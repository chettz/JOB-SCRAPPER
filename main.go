package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct{
	id			string
	title		string
	date		string
	condition 	string
	corp 		string
	
}

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?=&searchword=python"


func main() {
	totalPages := getPages()

	for i:=0;i<totalPages;i++{
		getPage(i+1)
	}
}

func getPage(page int){
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
		locations := []string{}
		id, _ := card.Attr("value")
		title := card.Find(".job_tit>a").Text()
		card.Find(".job_condition span").First().Find("a").Each(func(i int, s *goquery.Selection){
			locations = append(locations, s.Text())
		})
		
		fmt.Println(id, title, locations)

	})

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
	if res.StatusCode != 200{
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}