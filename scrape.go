package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

func GetTitle(url string) string {
	return url[6:]
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

func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "/wiki/") && !strings.Contains(url, ":") && !strings.Contains(url, "#") && !strings.Contains(url, "#")
}

func CheckLinks(url string) {
	var doc *goquery.Document
	var title string

	doc, title = loadHTML(url)

	fmt.Printf("Page title: %s\n\n", title)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && IsValidURL(link) {
			fmt.Println(link)
		}
	})
}

func main() {
	// Read starting page
	var startURL string = AskTitleInput("Input starting page: ")

	// Read ending page
	// var endURL string = AskTitleInput("Input ending page: ")

	// Get document and title
	var doc *goquery.Document
	// var title string

	doc, _ = loadHTML(startURL)

    // Open or create a file for writing
    file, err := os.OpenFile("result.txt", os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer file.Close() // Ensure the file is closed when the function returns

	// Find all links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && IsValidURL(link) {
			var newURL string = "https://en.wikipedia.org" + link
			
			_, err = io.WriteString(file, newURL + "\n")
		}
	})
}