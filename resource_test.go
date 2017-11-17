package main

import (
	"reflect"
	"testing"
)

func TestParseHTML(t *testing.T) {
	body := `<html><head><link rel="stylesheet" href="//a.com/test.css"><script src="/test.js"></script></head><body><img src="https://www.something.com/a.jpg"></a>`
	resources, err := extractResources("https", "test.com", ([]byte)(body))
	if err != nil {
		t.Fatalf("Parse failed with %v", err)
	}
	expected := []*resource{
		&resource{url: "https://a.com/test.css", resType: cssResource},
		&resource{url: "https://test.com/test.js", resType: scriptResource},
		&resource{url: "https://www.something.com/a.jpg", resType: imageResource},
	}
	if !reflect.DeepEqual(resources, expected) {
		t.Fatalf("Expected %v, but got %v", expected, resources)
	}
}
