package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type resourceType int

const (
	notRemoteResource resourceType = iota
	htmlResource
	cssResource
	scriptResource
	imageResource
)

type resource struct {
	resType resourceType
	url     string
	size    int
	err     error
}

func (rt resourceType) String() string {
	switch rt {
	case notRemoteResource:
		return "not a remote resource"
	case htmlResource:
		return "html"
	case cssResource:
		return "css"
	case scriptResource:
		return "script"
	case imageResource:
		return "img"
	default:
		return fmt.Sprintf("unknown resource type %d", rt)
	}
}

func extractResources(scheme, host string, body []byte) ([]*resource, error) {
	buff := bytes.NewBuffer(body)
	doc, err := html.Parse(buff)
	if err != nil {
		return nil, err
	}
	return parseResources(scheme, host, doc), nil
}

func parseResources(scheme, host string, node *html.Node) []*resource {
	resources := []*resource{}
	var ref resourceRef
	switch node.DataAtom {
	case atom.Link:
		ref = (*linkNode)(node)
	case atom.Img:
		ref = (*imgNode)(node)
	case atom.Script:
		ref = (*scriptNode)(node)
	}
	if ref != nil {
		if url, resType := ref.reference(); resType != notRemoteResource {
			if strings.HasPrefix(url, "//") {
				url = scheme + ":" + url
			} else if strings.HasPrefix(url, "/") {
				url = scheme + "://" + host + url
			}
			resources = append(resources, &resource{url: url, resType: resType})
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		resources = append(resources, parseResources(scheme, host, c)...)
	}
	return resources
}

func (r *resource) get() ([]byte, error) {
	resp, err := http.Get(r.url)
	if err != nil {
		r.err = err
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.err = err
		return nil, err
	}

	r.size = len(body)
	return body, nil
}

func (r *resource) String() string {
	return fmt.Sprintf("(%v|%v|%v|%v)", r.resType, r.url, r.size, r.err)
}
