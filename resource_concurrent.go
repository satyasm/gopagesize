package main

type request struct {
	res  *resource
	resp chan<- *result
}

type result struct {
	res  *resource
	body []byte
	err  error
}

func worker(reqChan <-chan *request) {
	for {
		if r, ok := <-reqChan; ok {
			body, err := r.res.get()
			r.resp <- &result{res: r.res, body: body, err: err}
		}
	}
}
