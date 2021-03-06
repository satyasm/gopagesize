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
	flag.Usage = func() {
		fmt.Printf("%s [flags] url [url...]\n", os.Args[0])
		fmt.Println("flags:")
		flag.PrintDefaults()
	}
	concurrent := flag.Bool("c", false, "specify to run in concurrent mode")
	poolSize := flag.Int("p", 10, "pool size when using concurrent mode")
	fileForUrls := flag.String("f", "", "file name to load list of url's from")
	slowestOnly := flag.Bool("s", false, "print only the slowest resource")
	resourceConcurrency := flag.Bool("r", false, "do resource level concurrency")
	hostConcurrency := flag.Bool("H", false, "host level concurrency")
	trace := flag.Bool("t", false, "trace network connections")
	flag.Parse()

	if len(flag.Args()) == 0 && *fileForUrls == "" {
		flag.Usage()
		os.Exit(1)
	}

	urls := flag.Args()
	if *fileForUrls != "" {
		if fileUrls, err := urlsFromFile(*fileForUrls); err == nil {
			urls = append(urls, fileUrls...)
		}
	}

	log.Printf("Using %d CPUs", runtime.GOMAXPROCS(-1))
	if *trace {
		startTrace()
	}

	startTime := time.Now()
	pages = resolveUrls(urls, *poolSize, *concurrent, *resourceConcurrency, *hostConcurrency)
	endTime := time.Now()
	printStats(pages, *concurrent, *slowestOnly)

	if *trace {
		fmt.Println()
		writeConnTrace(os.Stdout)
		fmt.Println()
	}
	log.Printf("Total time taken: %v", endTime.Sub(startTime))
}

func resolveUrls(urls []string, poolSize int, concurrent, resourceConcurrency, hostConcurrency bool) []*page {
	if concurrent {
		if resourceConcurrency {
			if hostConcurrency {
				return resolveConcurrentlyByHost(urls, poolSize)
			}
			return resolveConcurrently(urls, poolSize)
		}
		return resolveInGoRoutine(urls)
	}
	return resolveSynchronously(urls)
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

func printStats(pages []*page, concurrent bool, slowest bool) {
	tot := &stat{url: "Total"}
	writeStatsHeader(os.Stdout)
	fmt.Println()
	for _, p := range pages {
		s := p.stat(slowest)
		addTotal(tot, s, concurrent)
		s.write(os.Stdout)
		fmt.Println()
	}
	tot.write(os.Stdout)
}
