package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var pages []*page
	concurrent := flag.Bool("c", false, "specify to run in concurrent mode")
	p := flag.Int("p", 10, "pool size when using concurrent mode")
	f := flag.String("f", "", "file name to load list of url's from")
	flag.Parse()

	if len(flag.Args()) == 0 && *f == "" {
		fmt.Printf("%s [flags] url [url...]\n", os.Args[0])
		fmt.Println("flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	urls := flag.Args()
	if *f != "" {
		if fileUrls, err := urlsFromFile(*f); err == nil {
			urls = append(urls, fileUrls...)
		}
	}

	startTime := time.Now()
	if !*concurrent {
		pages = resolveSynchronously(urls)
	} else {
		pages = resolveConcurrently(urls, *p)
	}
	endTime := time.Now()

	for _, p := range pages {
		fmt.Printf("%-40s, %4d, %10d, %v\n", p.url, p.numResources(), p.total, p.timeTaken)
	}
	log.Printf("Total time taken: %v", endTime.Sub(startTime))
}

func urlsFromFile(fname string) ([]string, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	urls := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	return urls, nil
}

func resolveSynchronously(urls []string) []*page {
	pages := []*page{}
	for _, url := range urls {
		p, err := newPage(url)
		if err != nil {
			log.Printf("Error constructing page: %v", err)
			continue
		}
		if err := p.resolve(); err != nil {
			log.Printf("Error resolving page: %v", err)
		}
		pages = append(pages, p)
	}
	return pages
}

func resolveConcurrently(urls []string, nPoolSize int) []*page {
	pages := []*page{}
	reqChan := make(chan *request)
	for i := 0; i < nPoolSize; i++ {
		go worker(reqChan)
	}

	pagesChan := make(chan *page)
	numPages := 0
	for _, url := range urls {
		p, err := newPage(url)
		if err != nil {
			log.Printf("Error constructing page: %v", err)
			continue
		}
		numPages++
		go p.resolveConcurrently(reqChan, pagesChan)
	}
	for i := 0; i < numPages; i++ {
		if p, ok := <-pagesChan; ok {
			pages = append(pages, p)
		}
	}
	return pages
}
