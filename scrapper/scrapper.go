package scrapper

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)

type jazzItem struct {
	title string
	date  string
	url   string
}

func Scrapper(term string) {
	var baseUrl string = "https://www.jazzedmagazine.com/"
	var items []jazzItem
	extractC := make(chan []jazzItem)
	totalPages := getLastPage(baseUrl, term)
	fmt.Println("getLastPage: ", totalPages)

	for i := 1; i < totalPages; i++ {
		go getPage(i, extractC, baseUrl, term)

	}

	for i := 1; i < totalPages; i++ {
		extractedItems := <-extractC
		items = append(items, extractedItems...)
	}
	fmt.Println("Done, extracted ", len(items))

	writeItems(items)
	fmt.Println("Done Wrote :", len(items))
}

func getPage(page int, mainC chan<- []jazzItem, baseUrl string, term string) {
	var items []jazzItem
	c := make(chan jazzItem)
	pageUrl := baseUrl + "page/" + strconv.Itoa(page) + "/?s=" + term
	fmt.Println("Requesting: ", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkStatus(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchArticles := doc.Find(".archive>li")
	searchArticles.Each(func(i int, s *goquery.Selection) {
		go extractArticles(s, c)
	})

	for i := 0; i < searchArticles.Length(); i++ {
		item := <-c
		items = append(items, item)
	}
	mainC <- items
}

func extractArticles(s *goquery.Selection, c chan<- jazzItem) {
	title := CleanString(s.Find("h2").Text())
	date := CleanString(s.Find("span").Text())
	url, _ := s.Find("a").Attr("href")

	c <- jazzItem{
		title: title,
		date:  date,
		url:   url,
	}
}

func getLastPage(baseUrl, term string) int {
	pages := 0
	res, err := http.Get(baseUrl + "?s=" + term)
	checkErr(err)
	checkStatus(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		lastAHref, _ := s.Find("a").Last().Attr("href")
		reg, _ := regexp.Compile("[^0-9]+")
		pagesString := reg.ReplaceAllString(lastAHref, "")
		pages, _ = strconv.Atoi(pagesString)
	})

	return pages
}

func writeItems(items []jazzItem) {
	csv, err := ccsv.NewCsvWriter("jazz.csv")
	checkErr(err)

	defer csv.Close()

	csv.Write([]string{"Title", "Date", "Url"})

	done := make(chan bool)

	for _, item := range items {
		go func(item jazzItem) {
			csv.Write([]string{item.title, item.date, item.url})
			done <- true
		}(item)
	}
	for i := 0; i < len(items); i++ {
		<-done
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
func checkStatus(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status : ", res.StatusCode)
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
