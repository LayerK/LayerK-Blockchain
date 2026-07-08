package main

import (
	"strings"
	"testing"
)

func TestReadLimitedRPCResponseAllowsExactLimit(t *testing.T) {
	previousLimit := maxRPCResponseBytes
	maxRPCResponseBytes = 4
	t.Cleanup(func() {
		maxRPCResponseBytes = previousLimit
	})

	body, err := readLimitedRPCResponse(strings.NewReader("1234"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "1234" {
		t.Fatalf("unexpected response body %q", string(body))
	}
}

func TestReadLimitedRPCResponseRejectsOversizedBody(t *testing.T) {
	previousLimit := maxRPCResponseBytes
	maxRPCResponseBytes = 4
	t.Cleanup(func() {
		maxRPCResponseBytes = previousLimit
	})

	if _, err := readLimitedRPCResponse(strings.NewReader("12345")); err == nil {
		t.Fatal("expected oversized response error")
	}
}
