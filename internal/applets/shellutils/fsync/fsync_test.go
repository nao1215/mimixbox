package fsync

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

func TestFsyncFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t, f); err != nil {
		t.Errorf("fsync of an existing file should succeed, got %v", err)
	}
}

func TestFsyncMultiple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var files []string
	for _, n := range []string{"a", "b"} {
		f := filepath.Join(dir, n)
		if err := os.WriteFile(f, []byte(n), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, f)
	}
	if err := run(t, files...); err != nil {
		t.Errorf("fsync of two files should succeed, got %v", err)
	}
}

func TestFsyncMissing(t *testing.T) {
	t.Parallel()
	if err := run(t, "/no/such/fsync/file"); err == nil {
		t.Errorf("fsync of a missing file should fail")
	}
}

func TestFsyncNoOperand(t *testing.T) {
	t.Parallel()
	if err := run(t); err == nil {
		t.Errorf("fsync with no operand should fail")
	}
}
