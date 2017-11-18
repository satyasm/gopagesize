/*
GOPAGESIZE

gopagesize is a small program to explore various concurrency options with the Go programming
language. This program downloads a list of web pages, either given as command line parameters
or listed in a file (one URL, per line and specified with the -f option).

For each of the web page URL seen, it downloads the HTML, parses it to find all the link,
script and image tags and if these point to external resources, downloads them too. It them
computes the size of the web page plus the resources it references to give a sense of the total
page weight.

Usage

		gopagesize [flags] url [url...]
		flags:
		-H    host level concurrency
		-c    specify to run in concurrent mode
		-f string
		      file name to load list of url's from
		-p int
		      pool size when using concurrent mode (default 10)
		-r    do resource level concurrency
		-s    print only the slowest resource
		-t    trace network connections

Running it with just the URL's downloads each web page, parses it and then downloads each
resource found synchronously, one after the other. This is the basic mode with no concurrency
whatsoever.

The -c option turns on the concurrent mode. With just -c, each URL is processed in it's own
go routine, but the download, parse and fetch of resources happens synchronously per URL
just as in the case of the synchronous mode.

The -r option, along with -c, turns on request level concurrency. If specified, then a pool
of go routines are created, the pool size determined by the -p option, and all web requests,
including for the original web page and for each of the resources parsed then happens
concurrently within this pool. Like above, there is a go routine spawned per input web page
URL, which first fetches the webpage concurrently using the pool, parses it and then fetches
the resources concurrently, using the same pool.

The -H option, along with -c and -r, turns on per-host concurrency, which basically means
that a pool of go routines are maintained per host, instead of globally. The behavior is
similar to above, except that when the request is handed off by the go routine per webpage,
it goes to a router go routine when then multiplexes it to the appropriate pool based on the
host name in the request, instead of going directly to the pool. The pool size is again
controlled by the -p option. [Note:yes it's a capital H for this option, because the small
-h is already taken by convention to mean print-help]

The -s and -t options control how the results are printed. If -s is specified, then for
each webpage, in addition to the details of fetching the main page and the parse time,
one the slowest resource is printed. This is useful in concurrent mode to check if the
total time taken for the page is close to the sum of the time for the page, parse time
and the time to fetch the slowest resource, thus showing true concurrency.

The -t option, when specified, turns on request tracing and prints a summary of number of
connections made by host, along with the address resolved for the host. This is useful to
see the difference in connection behavior with and without the -H option, when fetching
concurrently.
*/
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
