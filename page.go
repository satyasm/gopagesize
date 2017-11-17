package main

import (
	"bytes"
	"fmt"
	"net/url"
	"time"
)

type page struct {
	url       *url.URL
	base      *resource
	assets    map[resourceType]map[string]*resource
	total     int
	timeTaken time.Duration
	parseTime time.Duration
	err       error
}

func newPage(rawURL string) (*page, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &page{
		url:    url,
		base:   &resource{url: rawURL, resType: htmlResource},
		assets: map[resourceType]map[string]*resource{},
	}, nil
}

func (p *page) numResources() (nResources int) {
	for _, a := range p.assets {
		nResources += len(a)
	}
	return
}

func (p *page) parseResources(body []byte) error {
	startTime := time.Now()
	defer func() {
		p.parseTime = time.Since(startTime)
	}()
	resources, err := extractResources(p.url.Scheme, p.url.Host, body)
	if err != nil {
		return err
	}
	for _, r := range resources {
		if _, found := p.assets[r.resType]; found {
			p.assets[r.resType][r.url] = r
		} else {
			p.assets[r.resType] = map[string]*resource{r.url: r}
		}
	}
	return nil
}

func (p *page) String() string {
	var buff bytes.Buffer
	fmt.Fprintf(&buff, "Page: %s %d\n", p.base.url, p.base.size)
	for t, a := range p.assets {
		fmt.Fprintf(&buff, "%v:\n", t)
		for _, r := range a {
			fmt.Fprintf(&buff, "  %s %d\n", r.url, r.size)
		}
	}
	fmt.Fprintf(&buff, "%s|Total = %d bytes\n", p.url, p.total)
	return buff.String()
}

func (p *page) slowest() (res *resource) {
	var maxTime time.Duration
	for _, a := range p.assets {
		for _, r := range a {
			if r.timeTaken > maxTime {
				maxTime = r.timeTaken
				res = r
			}
		}
	}
	return
}

func (p *page) stat(onlySlowest bool) *stat {
	s := &stat{
		url:         p.base.url,
		numRequests: p.numResources() + 1,
		size:        p.total,
		timeTaken:   p.timeTaken,
	}
	s.addComponent(&stat{url: p.base.url, numRequests: 1, size: p.base.size, timeTaken: p.base.timeTaken})
	s.addComponent(&stat{url: "parse", numRequests: 0, size: 0, timeTaken: p.parseTime})
	if onlySlowest {
		slowest := p.slowest()
		s.addComponent(&stat{url: slowest.url, numRequests: 1, size: slowest.size, timeTaken: slowest.timeTaken})
	} else {
		for _, a := range p.assets {
			for _, r := range a {
				s.addComponent(&stat{url: r.url, numRequests: 1, size: r.size, timeTaken: r.timeTaken})
			}
		}
	}
	return s
}
