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

	start := time.Now()
	if !*concurrent {
		resolveSynchronously(urls)
	} else {
		resolveConcurrently(urls, *p)
	}

	log.Printf("Total time taken: %v", time.Since(start))
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

func resolveSynchronously(urls []string) {
	for _, url := range urls {
		p, err := newPage(url)
		if err != nil {
			log.Printf("Error constructing page: %v", err)
			continue
		}
		if err := p.resolve(); err != nil {
			log.Printf("Error resolving page: %v", err)
			continue
		}
		fmt.Println(p)
	}
}

func resolveConcurrently(urls []string, nPoolSize int) {
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
			fmt.Println(p)
		}
	}
}
