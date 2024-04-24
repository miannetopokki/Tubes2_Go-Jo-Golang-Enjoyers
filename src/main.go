package main

import (
	"fmt"
	"net/http"
	"text/template"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseFiles("main.html"))
}

type wikiGameInfo struct {
	Source      string
	Destination string
}

type resultStruct struct {
	Path    []string
	Degrees int
	Time    int
	Artikel int
}

func WikiGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	infoSrcDest := wikiGameInfo{
		Source:      r.FormValue("src"),
		Destination: r.FormValue("dest"),
	}

	algorithm := r.FormValue("algorithm")

	succeed := true
	validSrc := false
	validDest := false
	sent := true
	result := fmt.Sprintf("%s -> %s", infoSrcDest.Source, infoSrcDest.Destination)

	srcLink := fmt.Sprintf("https://en.wikipedia.org/wiki/%s", infoSrcDest.Source)
	destLink := fmt.Sprintf("https://en.wikipedia.org/wiki/%s", infoSrcDest.Destination)

	if isValidWikiLink(srcLink) {
		validSrc = true
	}
	if isValidWikiLink(destLink) {
		validDest = true
	}

	tmpl.Execute(w, struct {
		Sent      bool
		Success   bool
		ValidSrc  bool
		ValidDest bool
		Results   wikiGameInfo
		Result    string
		Algorithm string
	}{sent, succeed, validSrc, validDest, infoSrcDest, result, algorithm})
}

func isValidWikiLink(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func main() {
	http.HandleFunc("/", WikiGame)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	link := "http://localhost:8080"
	fmt.Println("Server dimulai pada ", link)
	http.ListenAndServe(":8080", nil)
}
