package main

import "testing"

func TestPlaceholder(t *testing.T) {
	if 1 != 1 {
		t.Error("1 does not equal 1")
	}
}
