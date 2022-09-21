package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type jazzItem struct {
	title string
	date  string
	url   string
}

var baseUrl string = "https://www.jazzedmagazine.com/"

func main() {
	var items []jazzItem
	totalPages := 6 //getLastPage()
	for i := 1; i < totalPages; i++ {
		extractedItems := getPage(i)
		items = append(items, extractedItems...)

	}
	writeItems(items)
	fmt.Println("Done, extracted ", len(items))
}

func writeItems(items []jazzItem) {
	file, err := os.Create("jazz.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Date", "Url"}

	errWriteHeader := w.Write(headers)
	checkErr(errWriteHeader)

	for _, item := range items {
		itemSlice := []string{item.title, item.date, item.url}
		errWriteItems := w.Write(itemSlice)
		checkErr(errWriteItems)

	}
}

func getPage(page int) []jazzItem {
	var items []jazzItem
	pageUrl := baseUrl + "page/" + strconv.Itoa(page) + "/?s="
	fmt.Println("Requesting: ", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkStatus(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".archive>li").Each(func(i int, s *goquery.Selection) {
		item := extractMusic(s)
		items = append(items, item)
	})
	return items
}

func extractMusic(s *goquery.Selection) jazzItem {
	title := cleanString(s.Find("h2").Text())
	date := cleanString(s.Find("span").Text())
	url, _ := s.Find("a").Attr("href")

	return jazzItem{
		title: title,
		date:  date,
		url:   url,
	}
}

func getLastPage() int {
	pages := 0
	res, err := http.Get(baseUrl + "?s=")
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

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
