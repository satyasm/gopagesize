package main

import "log"

func resolvePage(result chan<- *page, p *page) {
	p.resolve()
	result <- p
}

func resolveInGoRoutine(urls []string) []*page {
	pages := []*page{}
	results := make(chan *page)
	numPages := 0
	for _, url := range urls {
		p, err := newPage(url)
		if err != nil {
			log.Printf("Error constructing page: %v", err)
			continue
		}
		numPages++
		go resolvePage(results, p)
	}
	for i := 0; i < numPages; i++ {
		if p, ok := <-results; ok {
			pages = append(pages, p)
		}
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
