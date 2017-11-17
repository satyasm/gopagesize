package main

import (
	"time"
)

func (p *page) resolveConcurrently(reqChan chan<- *request, pagesChan chan<- *page) {
	startTime := time.Now()
	pageRespChan := make(chan *result)
	reqChan <- &request{res: p.base, resp: pageRespChan}
	resp := <-pageRespChan
	if resp.err != nil {
		p.err = resp.err
		pagesChan <- p
		return
	}

	if err := p.parseResources(resp.body); err != nil {
		p.err = err
		pagesChan <- p
		return
	}
	p.total += p.base.size

	p.resolveResourcesConcurrently(reqChan)
	p.timeTaken = time.Since(startTime)
	pagesChan <- p
}

func (p *page) resolveResourcesConcurrently(reqChan chan<- *request) {
	nResources := p.numResources()
	resourcesChan := make(chan *result, nResources)

	for _, a := range p.assets {
		for _, r := range a {
			reqChan <- &request{res: r, resp: resourcesChan}
		}
	}

	for i := 0; i < nResources; i++ {
		resp := <-resourcesChan
		p.total += resp.res.size
	}
}
