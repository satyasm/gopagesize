package main

import (
	"fmt"
	"io"
	"sort"
	"time"
)

const urlLength = 50

var (
	mainFmt = fmt.Sprintf("%%-%ds, %%6d, %%13d, %%v\n", urlLength)
	compFmt = fmt.Sprintf("    %%-%ds, %%6d, %%13d, %%v\n", urlLength-4)
	hdrFmt  = fmt.Sprintf("%%-%ds, %%6s, %%13s, %%s\n", urlLength)
)

type stat struct {
	url         string
	timeTaken   time.Duration
	numRequests int
	size        int
	components  []*stat
}

func (s *stat) addComponent(c *stat) {
	s.components = append(s.components, c)
}

func stringMaxLen(length int, s string) string {
	if len(s) < length {
		return s
	}
	prefix := length/2 - 2
	suffix := length/2 - 1
	if length%2 == 1 {
		prefix++ // if we have odd length, we can include an extra char in prefix
	}
	return s[:prefix] + "..." + s[len(s)-suffix:]
}

func (s *stat) write(w io.Writer) {
	if len(s.components) > 2 {
		resources := s.components[2:] // by convention, first two are the main page and parse time
		// put the slowest resource on top
		sort.Slice(resources, func(i, j int) bool { return resources[i].timeTaken > resources[j].timeTaken })
	}
	fmt.Fprintf(w, mainFmt, stringMaxLen(urlLength, s.url), s.numRequests, s.size,
		s.timeTaken.Round(time.Millisecond))
	for _, c := range s.components {
		fmt.Fprintf(w, compFmt, stringMaxLen(urlLength-4, c.url), c.numRequests, c.size,
			c.timeTaken.Round(time.Millisecond))
	}
}

func writeStatsHeader(w io.Writer) {
	fmt.Fprintf(w, hdrFmt, "URL", "# res", "size (bytes)", "time taken")
}
