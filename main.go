package main

import (
	"os"
	"strings"

	"github.com/go_scrapper/scrapper"
	"github.com/labstack/echo"
)

var fileName string = "jazz.csv"

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handelScrape(c echo.Context) error {
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrapper(term)
	return c.Attachment(fileName, fileName)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handelScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
