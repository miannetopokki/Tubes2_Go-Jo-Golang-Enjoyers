package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func URLExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println(err)
		return false
	}
	// defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}

func CreateWikiURL(title string) string {
	return "https://en.wikipedia.org/wiki/" + title 
}

func AskTitleInput(prompt string) string {
	var title string
	
	// Get title
	fmt.Print(prompt)
	fmt.Scanln(&title)

	// Check if page exists
	url := CreateWikiURL(title)
	if !URLExists(url) {
		log.Fatal("Wikipedia page doesn't exist!")
	}
	return url
}

func MakeGETRequest(url string) *http.Response {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	// defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Status Code Error: %d, Status: %s", res.StatusCode, res.Status)
	}
	
	return res
}
func getHTMLDocument(res *http.Response) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	return doc
}

func loadHTML(url string) (*goquery.Document, string) {
	// Make HTTP request
	var res *http.Response = MakeGETRequest(url)
	
	// Get HTML document
	var doc *goquery.Document = getHTMLDocument(res)	

	// Return document and title
	var title string = doc.Find(".mw-page-title-main").First().Text()
	return doc, title
} 

func main() {
	// Read starting page
	var startURL string = AskTitleInput("Input starting page: ")

	// Read ending page
	// var endURL string = AskTitleInput("Input ending page: ")

	// Get document and title
	var doc *goquery.Document
	var title string

	doc, title = loadHTML(startURL)

	// Print title
	doc.Find("a")
	fmt.Println("Page title:", title)
}
