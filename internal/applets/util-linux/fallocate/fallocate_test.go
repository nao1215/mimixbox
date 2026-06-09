package fallocate

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestParseSize(t *testing.T) {
	t.Parallel()
	cases := map[string]int64{"0": 0, "1024": 1024, "1K": 1024, "2M": 2 * 1024 * 1024, "1G": 1024 * 1024 * 1024}
	for in, want := range cases {
		got, err := parseSize(in)
		if err != nil || got != want {
			t.Errorf("parseSize(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "abc", "-5", "1X2"} {
		if _, err := parseSize(bad); err == nil {
			t.Errorf("parseSize(%q) should fail", bad)
		}
	}
}

func size(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

func TestFallocateCreates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "new.bin")
	if err := run(t, "-l", "2048", f); err != nil {
		t.Fatalf("fallocate error = %v", err)
	}
	if size(t, f) != 2048 {
		t.Errorf("size = %d, want 2048", size(t, f))
	}
}

func TestFallocateNeverShrinks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(f, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t, "-l", "10", f); err != nil {
		t.Fatalf("fallocate error = %v", err)
	}
	if size(t, f) != 100 {
		t.Errorf("size = %d, want 100 (unchanged)", size(t, f))
	}
}

func TestFallocateErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "x")
	if err := run(t, f); err == nil {
		t.Errorf("missing -l should fail")
	}
	if err := run(t, "-l", "1024"); err == nil {
		t.Errorf("missing file should fail")
	}
	if err := run(t, "-l", "bad", f); err == nil {
		t.Errorf("invalid length should fail")
	}
}
