package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/allegro/bigcache"
)

var (
	cache *bigcache.BigCache
)
var uniqueLinkCount int
var visitedLinks = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

var destinationFound int32 = 0
var stopSearchClosed int32
var goroutineLimit int = 5 // pasang max goroutine disini
var reachedDestination bool = false

// func main() {
// 	// http://localhost:6060/debug/pprof/
// 	cache.items = make(map[string]*cachedItem)
// 	cache.maxItems = 1000 // Set batasan ukuran cache di sini

// 	go func() {
// 		log.Println(http.ListenAndServe("localhost:6060", nil))
// 	}()

// 	searchIDS("Indonesia", "Sun", 10)

// }

func searchIDS(source_link string, destination_link string, maxdepth int) resultStruct {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		log.Fatal(err)
	}

	// Redirect log output to ioutil.Discard to suppress log messages
	log.SetOutput(ioutil.Discard)
	var path []string
	var final_path []string
	degree := 0
	waktu := 0
	url := "https://en.wikipedia.org/wiki/" + source_link
	start := time.Now()

	for i := 0; i < maxdepth; i++ {
		fmt.Println("Searching in depth ", i+1, "...")
		dfs(source_link, url, destination_link, i, 0, &reachedDestination, &path, &final_path, cache)
		if reachedDestination {
			finish := time.Now()
			elapsed := finish.Sub(start)
			waktu = int(elapsed.Milliseconds())
			fmt.Println("Time : ", elapsed)
			fmt.Println("Total unique links:", uniqueLinkCount)
			fmt.Println("Path:", source_link, " -> ", strings.Join(final_path, " -> "))
			fmt.Println("Destination reached!")
			fmt.Println("Depth : ", i+1)
			degree = i + 1

			break
		} else {
			fmt.Println("Destination not found in depth ", i+1)
		}
	}
	reachedDestination = false
	result := resultStruct{
		Path:    final_path,
		Degrees: degree,
		Time:    waktu,
		Artikel: uniqueLinkCount,
	}
	uniqueLinkCount = 0
	visitedLinks.Lock()
	defer visitedLinks.Unlock()
	visitedLinks.m = make(map[string]bool)
	return result

}

func dfs(input string, url string, destination string, maxDepth int, currentDepth int, reachedDestination *bool, path *[]string, finalpath *[]string, cache *bigcache.BigCache) {

	if currentDepth > maxDepth || atomic.LoadInt32(&destinationFound) == 1 || *reachedDestination {
		return
	}

	doc, found := getFromCache(url, cache)
	if !found {
		var err error
		doc, err = goquery.NewDocument(url)
		if err != nil {
			log.Printf("Error loading %s: %v", url, err)
			return
		}
		cacheDocument(url, doc, cache)
	}

	var wg sync.WaitGroup
	stopSearch := make(chan struct{})
	var concurrencyLimit chan struct{}
	concurrencyLimit = make(chan struct{}, goroutineLimit)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && isValidLink(link, input, path) {
			visitedLinks.Lock()
			if !visitedLinks.m[url] {
				visitedLinks.m[url] = true
				visitedLinks.Unlock()
				uniqueLinkCount++
			} else {
				visitedLinks.Unlock()
			}
			newPath := append(*path, getTitleFromURL(link))
			// fmt.Println("Path:", input, " -> ", strings.Join(newPath, " -> "))

			if destination == getTitleFromURL(link) || link == "/wiki/"+destination || destination == removeChar(getTitleFromURL((link)), "_") {
				if !*reachedDestination {
					*path = newPath //Path success nemu
					*finalpath = append(*finalpath, input)
					*finalpath = append(*finalpath, newPath...)
					for idx, flink := range *finalpath {
						(*finalpath)[idx] = removeChar(flink, "_")
					}
				}
				*reachedDestination = true

				if atomic.LoadInt32(&stopSearchClosed) == 0 {
					close(stopSearch)
					atomic.StoreInt32(&stopSearchClosed, 1)
				}
				return
			}

			select {
			case <-stopSearch:
				return
			case concurrencyLimit <- struct{}{}:
				wg.Add(1)
				go func(link string) {
					defer func() {
						<-concurrencyLimit
						wg.Done()
					}()
					dfs(input, "https://en.wikipedia.org"+link, destination, maxDepth, currentDepth+1, reachedDestination, &newPath, finalpath, cache)
				}(link)
			}
		}
	})
	wg.Wait()
}

func getTitleFromURL(url string) string {
	if strings.HasPrefix(url, "/wiki/") {
		if idx := strings.Index(url, "#"); idx != -1 {
			url = url[:idx]
		}
		return strings.TrimPrefix(url, "/wiki/")
	}
	return ""
}
func removeChar(url string, c string) string {
	title := strings.TrimPrefix(url, "/wiki/")
	title = strings.ReplaceAll(title, c, " ")
	return title

}

func getFromCache(url string, cache *bigcache.BigCache) (*goquery.Document, bool) {
	entry, err := cache.Get(url)
	if err != nil {
		return nil, false
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(entry)))
	if err != nil {
		log.Printf("Error parsing cached document for %s: %v", url, err)
		return nil, false
	}
	return doc, true
}

func cacheDocument(url string, doc *goquery.Document, cache *bigcache.BigCache) {
	html, err := doc.Html()
	if err != nil {
		log.Printf("Error getting HTML content of document for %s: %v", url, err)
		return
	}
	if err := cache.Set(url, []byte(html)); err != nil {
		log.Printf("Error caching document for %s: %v", url, err)
	}
}

func isValidLink(link string, input string, path *[]string) bool {
	for _, visitedLink := range *path {
		if visitedLink == link {
			return false
		}
	}
	return strings.HasPrefix(link, "/wiki/") && !strings.Contains(link, ":") && !strings.Contains(link, "Main_Page") && getTitleFromURL(link) != input
}
