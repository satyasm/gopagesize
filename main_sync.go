package main

import "log"

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
