package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var cache struct {
	sync.Mutex
	items    map[string]*cachedItem
	maxItems int
}

//	type result struct{
//		path string[]
//		degrees int
//		time int
//		artikel int
//	}
type cachedItem struct {
	doc  *goquery.Document
	time time.Time
}

var uniqueLinkCount int
var visitedLinks = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

var destinationFound int32 = 0
var stopSearchClosed int32
var inProgress = make(map[string]struct{})
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
	var path []string
	var final_path []string
	degree := 0
	waktu := 0
	url := "https://en.wikipedia.org/wiki/" + source_link
	start := time.Now()

	for i := 0; i < maxdepth; i++ {
		fmt.Println("Searching in depth ", i+1, "...")
		dfs(source_link, url, destination_link, i, 0, &reachedDestination, &path, &final_path)
		if reachedDestination {
			finish := time.Now()
			elapsed := finish.Sub(start)
			waktu = int(elapsed.Seconds())
			fmt.Println("Time : ", elapsed)
			fmt.Println("Total unique links:", uniqueLinkCount)
			fmt.Println("Path:", source_link, " -> ", strings.Join(final_path, " -> "))
			fmt.Println("Destination reached!")
			fmt.Println("Depth : ", i+1)
			degree = i

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
	return result

}

func dfs(input string, url string, destination string, maxDepth int, currentDepth int, reachedDestination *bool, path *[]string, finalpath *[]string) {

	if currentDepth > maxDepth || atomic.LoadInt32(&destinationFound) == 1 || *reachedDestination {
		return
	}

	doc, found := getFromCache(url)
	if !found {
		var err error
		doc, err = goquery.NewDocument(url)
		if err != nil {
			log.Printf("Error loading %s: %v", url, err)
			return
		}
		cacheDocument(url, doc)
	}

	var wg sync.WaitGroup
	stopSearch := make(chan struct{})
	var concurrencyLimit chan struct{}
	concurrencyLimit = make(chan struct{}, goroutineLimit)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && isValidLink(link) {
			visitedLinks.Lock()
			if !visitedLinks.m[link] {
				visitedLinks.m[link] = true
				visitedLinks.Unlock()
				// fmt.Println(getTitleFromURL(link))
				uniqueLinkCount++
			} else {
				visitedLinks.Unlock()
			}
			newPath := append(*path, getTitleFromURL(link))
			// fmt.Println("Path:", input, " -> ", strings.Join(newPath, " -> "))

			if destination == getTitleFromURL(link) || link == "/wiki/"+destination {
				if !*reachedDestination {
					*path = newPath //Path success nemu
					*finalpath = newPath
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
					dfs(input, "https://en.wikipedia.org"+link, destination, maxDepth, currentDepth+1, reachedDestination, &newPath, finalpath)
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

func getFromCache(url string) (*goquery.Document, bool) {
	cache.Lock()
	defer cache.Unlock()
	item, found := cache.items[url]
	if found && time.Since(item.time) > time.Hour {
		delete(cache.items, url)
		found = false
	}
	if !found {
		return nil, false
	}
	return item.doc, true
}

func cacheDocument(url string, doc *goquery.Document) {
	cache.Lock()
	defer cache.Unlock()
	if len(cache.items) >= cache.maxItems {
		evictOldestFromCache()
	}
	cache.items[url] = &cachedItem{doc: doc, time: time.Now()}
}

func evictOldestFromCache() {
	oldestURL := ""
	oldestTime := time.Now()
	for url, item := range cache.items {
		if item.time.Before(oldestTime) {
			oldestURL = url
			oldestTime = item.time
		}
	}
	delete(cache.items, oldestURL)
}

func isValidLink(link string) bool {
	return strings.HasPrefix(link, "/wiki/") && !strings.Contains(link, ":") && !strings.Contains(link, "Main_Page")
}
