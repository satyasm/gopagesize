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
	reqChan := make(chan *request)

	// create pool of workers
	for i := 0; i < nPoolSize; i++ {
		go worker(reqChan)
	}

	return doConcurrently(reqChan, urls)
}

func doConcurrently(reqChan chan<- *request, urls []string) []*page {
	pages := []*page{}
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

func routeByHost(reqChan <-chan *request, nPoolSize int) {
	byHost := map[string]chan *request{}

	for {
		if r, ok := <-reqChan; ok {
			host := r.res.host()
			if _, found := byHost[host]; !found {
				c := make(chan *request)
				byHost[host] = c
				for i := 0; i < nPoolSize; i++ {
					go worker(c)
				}
			}
			byHost[host] <- r
		}
	}
}

func resolveConcurrentlyByHost(urls []string, nPoolSize int) []*page {
	reqChan := make(chan *request)
	go routeByHost(reqChan, nPoolSize)
	return doConcurrently(reqChan, urls)
}
