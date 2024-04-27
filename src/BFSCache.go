package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	cacheBFS map[string][]string = make(map[string][]string)
	cacheBFSMutex sync.RWMutex = sync.RWMutex{}
)

// Convert title to link
func TitleToLink(title string) string {
	return strings.Replace(title, " ", "_", -1);
}

// Convert link to title
func LinkToTitle(link string) string {
	return strings.Replace(link, "_", " ", -1);
}

// Read cache from file
func ReadCache() {
	// Open the cache file for reading
	file, err := os.OpenFile("cache.txt", os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()
	
	// Create a scanner to read from the file
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024*1024)
	
	// Process each line in the file
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		
		if len(parts) == 0 {
			continue // Skip empty lines
		}
		
		key := parts[0]
		values := parts[1:]
		cacheBFS[key] = append(cacheBFS[key], values...)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}

// Write cache to file
func WriteCache() {
	// Remove old cache file
	os.Remove("cache.txt")

	// Create new cache file
	file2, err := os.OpenFile("cache.txt", os.O_CREATE, 0644);
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file2.Close()

	// Write values from the map to the cache file
	for key, values := range cacheBFS {
		line := fmt.Sprintf("%s %s\n", key, strings.Join(values, " "))

		// Write the line to the file
		_, err := file2.WriteString(line)
		if err != nil {
			log.Fatalf("Error writing to file: %v", err)
		}
	}
}