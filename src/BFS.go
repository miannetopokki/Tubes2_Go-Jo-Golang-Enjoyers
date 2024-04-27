package main

// Imported libraries
import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Color constants (for printing purposes)
const (
    Reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
)

// Link struct that represents a Wikipedia article
type Link struct {
	title string	// Title of the article
	url string		// URL of the article
	path []string	// Path to the article from the starting article
	iter int		// Degree of separation
}

// Global variables
var (
	seen map[string]bool		// Map of seen articles
	found bool					// Flag to indicate if the end article has been found
	queue QueueLinked[Link]		// Queue of articles to visit
	startTitle string			// Title of the starting article
	endTitle string				// Title of the end article
	queueMutex sync.Mutex		// Mutex for the queue
	seenMutex sync.RWMutex		// Mutex for the seen map
	
    workerCount = 100 			// Number of concurrent workers
    workerSem   = make(chan struct{}, workerCount)	// Semaphore for workers

	articles int				// Number of articles visited
	articlesMutex sync.Mutex	// Mutex for the number of articles visited

	ResultPath 		[]string	// Resulting path from starting article to end article
	ResultDegrees 	int			// Degrees of separation between starting article and end article
	ResultTime 		int			// Time taken to find the path to the end article
	ResultArtikel 	int			// Number of articles visited
)

// Function to convert a title to a link
func CreateWikiURL(title string) string {
	return "https://en.wikipedia.org/wiki/" + title 
}

// Function to get the HTML document of an URL
func getHTMLDocument(url string) *goquery.Document {
	doc, err := goquery.NewDocument(url)
	
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

// Function to get the document and title of an URL
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

// Function to check if a given Wikipedia URL is valid (for the purposes of this program)
func IsValidURL(url string, title string) bool {
	return strings.HasPrefix(url, "/wiki/") && !strings.Contains(url, ":") && !strings.Contains(url, "#") && !strings.Contains(url, "#") && url != "/wiki/Main_Page" && title != "View the content page [c]"
}

// Procedure to write the currently checked article
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

// Function to check if the end article has been found
func CheckFound(title string, path []string, iter int) bool {
	// Convert path from list of string to string
	var pathString string = strings.Join(path, " -> ")
	
	// Check if given title is the end title
	if strings.EqualFold(title, endTitle) {
		// Print path and set flag
		pathString += " -> " + title
		fmt.Println(Green + "\nPath:\n" + Reset + pathString)
		found = true

		// Set result variables to send back to frontend code
		ResultPath = append(path, title)
		ResultDegrees = iter + 1
		ResultArtikel = articles
	}

	return found
}

// Procedure to iterate and check all links of a given Wikipedia article
func GetLinks(doc *goquery.Document, docTitle string, iter int, path []string) {
	// Initialize list of links
	var linkList []string = []string{}

	// Iterate over links
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		link, exists := s.Attr("href")
		title, _ := s.Attr("title")

		if idx := strings.Index(link, "#"); idx != -1 {
			link = link[:idx]
		}

		if exists && IsValidURL(link, title) {
			// Check if link is found
			if CheckFound(title, path, iter) {
				return false
			}

			// Add link to list
			linkList = append(linkList, TitleToLink(title))
	
			// Get full URL
			var newURL string = "https://en.wikipedia.org" + link

			// Add link to queue
			queueMutex.Lock()
			queue.Enqueue(Link{title, newURL, append(path, title), iter + 1})
			queueMutex.Unlock()
		}
		return true
	})

	// Add links to cache
	if !found {
		cacheBFSMutex.Lock()
		cacheBFS[TitleToLink(docTitle)] = linkList
		cacheBFSMutex.Unlock()
	}
}

// Procedure to iterate and check all links of a given Wikipedia article (from cache)
func GetLinksCache(parent string, iter int, path[] string) {
	// Get links from cache
	cacheBFSMutex.RLock()
	var links []string = cacheBFS[parent]
	cacheBFSMutex.RUnlock()

	// Iterate over links
	for _, link := range links {
		var title string = LinkToTitle(link)
		
		if CheckFound(title, path, iter) {
			return
		}

		// Get new URL
		var newURL string = CreateWikiURL(link)

		// Add link to queue
		queueMutex.Lock()	
		queue.Enqueue(Link{title, newURL, append(path, title), iter + 1})
		queueMutex.Unlock()
	}
}

// Procedure to start BFS traversal
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

		// Sleep for a small duration after sending HTTP request to avoid timeout
		if !ok {
			time.Sleep(22 * time.Millisecond)
		}
	}
}

// BFS wrapper function called by backend
func BFS(startPage string, endPage string) resultStruct {
	articles = 0
	var startURL string = CreateWikiURL(startPage)
	var endURL string = CreateWikiURL(endPage)
	
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

	// Calculate time taken
    duration := time.Since(start)
    seconds := duration.Seconds()
	ms := duration.Milliseconds()

	fmt.Println(Green + "\nTime taken: " + Yellow + strconv.FormatFloat(seconds, 'f', 6, 64) + " sec" + Reset)
	fmt.Println(Green + "Time taken: " + Yellow + strconv.FormatInt(ms, 10) + " ms" + Reset)

	// Return result
	return resultStruct{
		Path:    ResultPath,
		Degrees: ResultDegrees,
		Time:    int(ms),
		Artikel: ResultArtikel,
	}
}