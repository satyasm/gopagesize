package main

import "testing"

func TestStringMaxLen(t *testing.T) {
	src := "This is a really long string"

	with16 := stringMaxLen(16, src)
	expectedWith16 := "This i... string"
	if expectedWith16 != with16 {
		t.Errorf("Expected: %s, but got: %s", expectedWith16, with16)
	}

	with15 := stringMaxLen(15, src)
	expectedWith15 := "This i...string"
	if expectedWith15 != with15 {
		t.Errorf("Expected: %s, but got: %s", expectedWith15, with15)
	}
}
