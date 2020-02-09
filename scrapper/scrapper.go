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

type extractedJob struct {
	id       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

//Scrape scrapes the provided term
func Scrape(term string) {
	var baseURL string = "https://www.indeed.co.in/jobs?q=" + term + "&limit=20"
	var jobs []extractedJob
	totalPages := getPages(baseURL)
	channel := make(chan []extractedJob)
	//writeC := make(chan )
	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, channel)
	}
	for i := 0; i < totalPages; i++ {
		extractedJobs := <-channel
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPages(url string) int {
	pages := 0
	response, err := http.Get(url)
	checkError(err)
	checkCode(response)
	defer response.Body.Close()
	doc, err := goquery.NewDocumentFromReader(response.Body)
	checkError(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func getPage(page int, url string, channel chan<- []extractedJob) {
	var jobs []extractedJob
	pageURL := url + "&start=" + strconv.Itoa(page*20)
	c := make(chan extractedJob)
	fmt.Println("Requesting:", pageURL)
	res, err := http.Get(pageURL)
	checkError(err)
	checkCode(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	channel <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	company := CleanString(card.Find(".company").Text())
	location := CleanString(card.Find(".location").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- extractedJob{
		id:       id,
		title:    title,
		company:  company,
		location: location,
		salary:   salary,
		summary:  summary}
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkError(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Company", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkError(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://www.indeed.co.in/viewjob?jk=" + job.id, job.title, job.company, job.location, job.salary, job.summary}
		jwErr := w.Write(jobSlice)
		checkError(jwErr)

	}
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(response *http.Response) {
	if response.StatusCode != 200 {
		log.Fatalln("Response Faild with Status code:", response.StatusCode)
	}
}

//CleanString cleans the string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
