package main

import (
	"os"
	"strings"
	"testing"
)

// TestReadmeUpToDate fails when README.md's command list is stale, i.e. an
// applet was added, removed or had its synopsis changed without running
// `make command-list`. This keeps the documented list and the binary in sync.
func TestReadmeUpToDate(t *testing.T) {
	data, err := os.ReadFile("../../" + readmePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	start := strings.Index(content, startMark)
	end := strings.Index(content, endMark)
	if start < 0 || end < 0 || end < start {
		t.Fatalf("markers %q / %q not found in README.md", startMark, endMark)
	}

	got := content[start+len(startMark) : end]
	want := "\n" + table()
	if got != want {
		t.Errorf("README.md command list is out of date; run `make command-list`.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
