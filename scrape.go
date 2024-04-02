package main

import (
	"fmt"
	"log"
	"net/http"
)

func URLExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()

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
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Status Code Error: %d, Status: %s", res.StatusCode, res.Status)
	}
	
	return res
}

func main() {
	// Read starting page
	startURL := AskTitleInput("Input starting page: ")

	// Read ending page
	endURL := AskTitleInput("Input ending page: ")

	fmt.Println(startURL)
	fmt.Println(endURL)
	
	// Make HTTP GET request
	res := MakeGETRequest(startURL)
	fmt.Println(res)
}
