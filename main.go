package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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

	log.Printf("Using %d CPUs", runtime.GOMAXPROCS(-1))
	startTime := time.Now()
	if !*concurrent {
		pages = resolveSynchronously(urls)
	} else {
		pages = resolveConcurrently(urls, *p)
	}
	endTime := time.Now()
	printStats(pages, *concurrent)
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

func addTotal(s, c *stat, concurrent bool) {
	s.numRequests += c.numRequests
	s.size += c.size
	if concurrent {
		if s.timeTaken < c.timeTaken {
			s.timeTaken = c.timeTaken
		}
	} else {
		s.timeTaken += c.timeTaken
	}
}

func printStats(pages []*page, concurrent bool) {
	tot := &stat{url: "Total"}
	writeStatsHeader(os.Stdout)
	fmt.Println()
	for _, p := range pages {
		s := p.stat(false)
		addTotal(tot, s, concurrent)
		s.write(os.Stdout)
		fmt.Println()
	}
	tot.write(os.Stdout)
}
