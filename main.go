package main

import (
	"learngo/scrapper"
	"os"
	"strings"

	"github.com/labstack/echo"
)

const fileName string = "jobs.csv"

func handelHome(c echo.Context) error {
	return c.File("home.html")
}

func handelScrapper(c echo.Context) error {
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	defer os.Remove(fileName)
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)
}

func main() {
	//scrapper.Scrape("c")
	e := echo.New()
	e.GET("/", handelHome)
	e.POST("/scrape", handelScrapper)
	e.Logger.Fatal(e.Start(":1323"))
}
