package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestRegistryUpToDate fails when internal/applets/applet_registry_gen.go is
// stale, i.e. an applet package was added or removed without running
// `make generate`. This is the CI guard that the generated registry never
// drifts from the applet packages on disk.
func TestRegistryUpToDate(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}

	pkgs, err := scan(filepath.Join(root, appletsRel))
	if err != nil {
		t.Fatal(err)
	}
	got, err := render(pkgs)
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.Join(root, outputRel))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("%s is out of date; run `make generate`.", outputRel)
	}
}
