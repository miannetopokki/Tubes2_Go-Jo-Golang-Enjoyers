package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
    Reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
)

type Link struct {
	title string
	url string
	path []string
	iter int
}

var (
	seen map[string]bool
	found bool
	queue QueueLinked[Link]
	startTitle string
	endTitle string
	queueMutex sync.Mutex
	seenMutex sync.RWMutex
	
    workerCount = 100 // Number of concurrent workers
    workerSem   = make(chan struct{}, workerCount)

	articles int
	articlesMutex sync.Mutex

	ResultPath 		[]string
	ResultDegrees 	int
	ResultTime 		int
	ResultArtikel 	int
)

func CreateWikiURL(title string) string {
	return "https://en.wikipedia.org/wiki/" + title 
}

func getHTMLDocument(url string) *goquery.Document {
	doc, err := goquery.NewDocument(url)
	
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func loadHTML(url string) (*goquery.Document, string) {
	// Get HTML document
	var doc *goquery.Document = getHTMLDocument(url)	

	// Return document and title
	var title string = doc.Find(".mw-page-title-main").First().Text()
	var title_alt string = doc.Find(".firstHeading.mw-first-heading").First().Text()
	if title == "" {
		title = title_alt
	}
	return doc, title
}

func IsValidURL(url string, title string) bool {
	return strings.HasPrefix(url, "/wiki/") && !strings.Contains(url, ":") && !strings.Contains(url, "#") && !strings.Contains(url, "#") && url != "/wiki/Main_Page" && title != "View the content page [c]"
}

func WriteAndPrintRoot(iter int, title string, path []string) {
	// Print
	fmt.Print(Green)
	fmt.Printf("%d ", iter + 1)
	for _ = range iter {
		fmt.Printf("----")
	}

	var pathString string = strings.Join(path, " -> ")
	fmt.Println(" Root: " + Yellow + title + Reset + " (" + pathString + ")")
}

func CheckFound(title string, path []string, iter int) bool {
	var pathString string = strings.Join(path, " -> ")
	
	if strings.EqualFold(title, endTitle) {
		pathString += " -> " + title
		fmt.Println(Green + "\nPath:\n" + Reset + pathString)
		found = true

		ResultPath = append(path, title)
		ResultDegrees = iter + 1
		ResultArtikel = articles
	}

	return found
}

func GetLinks(doc *goquery.Document, docTitle string, iter int, path []string) {
	var linkList []string = []string{}

	// Iterate over links
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		link, exists := s.Attr("href")
		title, _ := s.Attr("title")

		if exists && IsValidURL(link, title) {
			// fileMutex.Lock()
			// file.WriteString("(scrape)" + title + " (" + strings.Join(path, " -> ") + ")" + "\n")
			// fileMutex.Unlock()
			// Check if link is found
			if CheckFound(title, path, iter) {
				return false
			}

			// Add link to list
			linkList = append(linkList, TitleToLink(title))
	
			// Get full URL
			var newURL string = "https://en.wikipedia.org" + link

			// Add link to queue
			// if heuristic[title] {
				// 	queue.EnqueueHead(Link{title, newURL, append(path, title), iter + 1})
				// } else {
			// }
			queueMutex.Lock()
			queue.Enqueue(Link{title, newURL, append(path, title), iter + 1})
			queueMutex.Unlock()
		}
		return true
	})

	if !found {
		cacheBFSMutex.Lock()
		cacheBFS[TitleToLink(docTitle)] = linkList
		cacheBFSMutex.Unlock()
	}
}

func GetLinksCache(parent string, iter int, path[] string) {
	// Get links from cache
	cacheBFSMutex.RLock()
	var links []string = cacheBFS[parent]
	cacheBFSMutex.RUnlock()

	// Iterate over links
	for _, link := range links {
		var title string = LinkToTitle(link)
		// fileMutex.Lock()
		// file.WriteString("(cache)" + title + " (" + strings.Join(path, " -> ") + ")" + "\n")
		// fileMutex.Unlock()
		
		if CheckFound(title, path, iter) {
			return
		}

		// Get new URL
		var newURL string = CreateWikiURL(link)

		// Add link to queue
		// if heuristic[title] {
			// 	queue.EnqueueHead(Link{title, newURL, append(path, title), iter + 1})
			// } else {
				// }
		queueMutex.Lock()	
		queue.Enqueue(Link{title, newURL, append(path, title), iter + 1})
		queueMutex.Unlock()
	}
}

func BFSTraversal() {
	// Start iteration	
	for !found {
		// Get queue head
		for queue.IsEmpty() {
			time.Sleep(10 * time.Millisecond)
		}
		queueMutex.Lock()
		var L Link = queue.Dequeue()
		queueMutex.Unlock()

		// Get queue head attributes
		var iter int = L.iter
		var title string = L.title
		var url string = L.url
		var path []string = L.path
		
		// Check if title is seen
		var hasSeen bool = false
		seenMutex.RLock()
		if seen[title] {
			hasSeen = true
		}
		seenMutex.RUnlock()

		if hasSeen || title == "" {
			continue
		}

		// Increment articles
		articlesMutex.Lock()
		articles++
		articlesMutex.Unlock()

		// If title == endTitle, end loop
		if CheckFound(title, path, iter) {
			return
		}
		
		// Write root link
		// fileMutex.Lock()
		// WriteAndPrintRoot(iter, title, path)
		// fileMutex.Unlock()
		
		// Set title to seen
		seenMutex.Lock()
		seen[title] = true
		seenMutex.Unlock()
		
		// Check if page is in cache
		cacheBFSMutex.RLock()
		_, ok := cacheBFS[TitleToLink(title)]
		cacheBFSMutex.RUnlock()
		
		// Acquire worker semaphore
		workerSem <- struct{}{}

		go func() {
			defer func() { <-workerSem }() // Release worker semaphore
	
			// If page is in cache, get links from cache
			if ok {
				fmt.Print(Yellow + "Cache hit: " + Reset)
				WriteAndPrintRoot(iter, title, path)
				GetLinksCache(TitleToLink(title), iter, path)
			} else {
				// Load HTML
				fmt.Print(Red + "Cache miss: " + Reset)
				WriteAndPrintRoot(iter, title, path)
				var doc *goquery.Document
				var docTitle string
				doc, docTitle = loadHTML(url)
		
				// Get links
				GetLinks(doc, docTitle, iter, path)
			}
		}()

		if !ok {
			time.Sleep(22 * time.Millisecond)
		}
	}
}
func BFS(startPage string, endPage string) resultStruct {
	articles = 0
	var startURL string = CreateWikiURL(startPage)
	var endURL string = CreateWikiURL(endPage)

	// file, _ = os.OpenFile("debug.txt", os.O_WRONLY|os.O_CREATE, 0644)
	
	// Read starting page
	_, startTitle = loadHTML(startURL)

	// Read ending page
	_, endTitle = loadHTML(endURL)

	// Initialize map of found pages
	seen = make(map[string]bool)

	// Initialize link queue
	queue = QueueLinked[Link]{}
	queue.Enqueue(Link{startTitle, startURL, []string{startTitle}, 0})
	
	// Initialize found flag
	found = false
	// initWorkers()

	// Instantly return if start == end
	if strings.EqualFold(startTitle, endTitle) {
		return resultStruct{
			Path:    []string{startTitle},
			Degrees: 0,
			Time:    0,
			Artikel: 0,
		}
	}

	// Start BFS traversal
	start := time.Now()
	BFSTraversal()

    duration := time.Since(start)
    seconds := duration.Seconds()
	ms := duration.Milliseconds()

	fmt.Println(Green + "\nTime taken: " + Yellow + strconv.FormatFloat(seconds, 'f', 6, 64) + " sec" + Reset)
	fmt.Println(Green + "Time taken: " + Yellow + strconv.FormatInt(ms, 10) + " ms" + Reset)

	return resultStruct{
		Path:    ResultPath,
		Degrees: ResultDegrees,
		Time:    int(ms),
		Artikel: ResultArtikel,
	}
}