package main

import "golang.org/x/net/html"

type linkNode html.Node
type scriptNode html.Node
type imgNode html.Node

type resourceRef interface {
	reference() (string, resourceType)
}

func (l *linkNode) reference() (url string, resType resourceType) {
	for _, a := range l.Attr {
		if a.Key == "rel" {
			switch a.Val {
			case "stylesheet":
				resType = cssResource
			case "shortcut icon":
				fallthrough
			case "apple-touch-icon-precomposed":
				fallthrough
			case "icon":
				resType = imageResource
			}
		}
		if a.Key == "href" {
			url = a.Val
		}
	}
	return
}

func (s *scriptNode) reference() (url string, resType resourceType) {
	for _, a := range s.Attr {
		if a.Key == "src" {
			resType = scriptResource
			url = a.Val
		}
	}
	return
}

func (i *imgNode) reference() (url string, resType resourceType) {
	for _, a := range i.Attr {
		if a.Key == "src" {
			resType = imageResource
			url = a.Val
		}
	}
	return
}
