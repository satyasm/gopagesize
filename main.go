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
	printStats(pages)
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

func stringMaxLen(length int, s string) string {
	if len(s) < length {
		return s
	}
	return s[:length-3] + "..."
}

func printStats(pages []*page) {
	fmt.Printf("%-40s, %6s, %13s, %s\n", "URL", "# res", "size (bytes)", "time taken")
	fmt.Println()
	for _, p := range pages {
		fmt.Printf("%-40s, %6d, %13d, %v\n", p.url, p.numResources(), p.total, p.timeTaken.Round(time.Millisecond))
		fmt.Printf("  %-38s, %6d, %13d, %v\n", "page", 1, p.base.size, p.base.timeTaken.Round(time.Millisecond))
		fmt.Printf("  %-38s, %6d, %13d, %v\n", "parse", 1, 0, p.parseTime.Round(time.Millisecond))
		r := p.slowest()
		fmt.Printf("  %-38s, %6d, %13d, %v\n", stringMaxLen(38, r.url), 1, r.size, r.timeTaken.Round(time.Millisecond))
		fmt.Println()
	}
}
