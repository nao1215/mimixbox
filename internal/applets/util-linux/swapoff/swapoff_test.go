package swapoff

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, swapoffErr error) *[]string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "swaps")
	content := "Filename\tType\tSize\tUsed\tPriority\n/dev/sda2 partition 2097148 0 -2\n/swapfile file 1024 0 -3\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	var called []string
	op, of := swapsPath, swapoffFn
	swapsPath = f
	swapoffFn = func(path string) error {
		called = append(called, path)
		return swapoffErr
	}
	t.Cleanup(func() { swapsPath, swapoffFn = op, of })
	return &called
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestDisableOne(t *testing.T) {
	called := setup(t, nil)
	if err := run(t, "/swapfile"); err != nil {
		t.Fatal(err)
	}
	if len(*called) != 1 || (*called)[0] != "/swapfile" {
		t.Errorf("swapoffFn called with %v", *called)
	}
}

func TestDisableAll(t *testing.T) {
	called := setup(t, nil)
	if err := run(t, "-a"); err != nil {
		t.Fatal(err)
	}
	got := append([]string{}, *called...)
	sort.Strings(got)
	if len(got) != 2 || got[0] != "/dev/sda2" || got[1] != "/swapfile" {
		t.Errorf("-a disabled %v, want both swaps", got)
	}
}

func TestDisableFails(t *testing.T) {
	setup(t, errors.New("operation not permitted"))
	if err := run(t, "/swapfile"); err == nil {
		t.Errorf("a disable failure should fail")
	}
}

func TestNoArg(t *testing.T) {
	setup(t, nil)
	if err := run(t); err == nil {
		t.Errorf("no target should fail")
	}
}

func TestDisableAllEmpty(t *testing.T) {
	// -a with no active swaps (empty table) is a no-op success, not an error.
	dir := t.TempDir()
	empty := filepath.Join(dir, "swaps")
	if err := os.WriteFile(empty, []byte("Filename\tType\tSize\tUsed\tPriority\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	op, of := swapsPath, swapoffFn
	swapsPath = empty
	swapoffFn = func(string) error { t.Fatal("should not be called"); return nil }
	defer func() { swapsPath, swapoffFn = op, of }()
	if err := run(t, "-a"); err != nil {
		t.Errorf("-a with no swaps should succeed, got %v", err)
	}
}
