// package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// func main() {
	// URL of the Wikipedia page you want to scrape
	url := "https://id.wikipedia.org/wiki/Institut_Teknologi_Bandung"
	url := "https://wikipedia.org/wiki/Institut_Teknologi_Bandung"

	// Make HTTP GET request
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load HTML document
	// doc, err := goquery.NewDocumentFromReader(res.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	fmt.Println(url)

	// Find the index of the keyword
	index := strings.Index(url, "/wiki/")
	if index == -1 {
		// Keyword not found
		fmt.Println("Keyword not found in the string.")
		return
	}

	// Get the substring starting from the index of the keyword
	substring := url[index:]
	fmt.Println("Substring starting from the keyword:", substring)

	// Find all <a> elements
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		// Get the href attribute value
		link, exists := s.Attr("href")
		if exists {
			// Filter out non-Wikipedia links
			if strings.HasPrefix(link, "/wiki/") && !strings.Contains(link, ":") && !strings.Contains(link, "#") && !strings.Contains(link, ".") {
				// Print the link
				fmt.Println(link)
			}
		}
	})
// }
