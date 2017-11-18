package main

import (
	"time"
)

func (p *page) resolveConcurrently(reqChan chan<- *request, pagesChan chan<- *page) {
	startTime := time.Now()
	defer func() {
		p.timeTaken = time.Since(startTime)
	}()
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
	pagesChan <- p
}

func (p *page) resolveResourcesConcurrently(reqChan chan<- *request) {
	// we create a buffered channel to receives the results from, with
	// the buffer size set to the number of resources to fetch. This is
	// important because the send reqChan is not assumed to be buffered,
	// this is routine first queues all the requests before starting
	// to read the response. Having the result channel buffered means
	// that the worker threads on reqChan will also be able to send a
	// result even if the reader is not ready, thus avoiding blocking
	// the workers. Consider the degenerate case of having only one
	// worker go routine receiving on reqChan. Without this buffer,
	// the whole process could hang, with writes blocked on reqChan
	// and the result chan.
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
