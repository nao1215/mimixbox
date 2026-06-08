package unzip_test

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/unzip"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := unzip.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// makeZip writes a zip archive containing the given name->content entries.
func makeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := unzip.New()
	if got := c.Name(); got != "unzip" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestExtract(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.zip")
	makeZip(t, archive, map[string]string{
		"a.txt":     "alpha",
		"sub/b.txt": "beta",
	})

	dest := filepath.Join(dir, "out")
	if _, errOut, err := run(t, "-d", dest, archive); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dest, "a.txt"))
	if err != nil {
		t.Fatalf("read a.txt: %v", err)
	}
	if string(got) != "alpha" {
		t.Errorf("a.txt = %q", got)
	}
	got, err = os.ReadFile(filepath.Join(dest, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("read sub/b.txt: %v", err)
	}
	if string(got) != "beta" {
		t.Errorf("sub/b.txt = %q", got)
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.zip")
	makeZip(t, archive, map[string]string{"only.txt": "hi"})

	out, _, err := run(t, "-l", archive)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "only.txt") {
		t.Errorf("list = %q", out)
	}
	// Listing must not create files.
	if _, statErr := os.Stat(filepath.Join(dir, "only.txt")); statErr == nil {
		t.Error("-l should not extract files")
	}
}

func TestUsageError(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Error("expected usage error without archive")
	}
	if !strings.Contains(errOut, "usage") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingArchive(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, filepath.Join(t.TempDir(), "nope.zip"))
	if err == nil {
		t.Error("expected error for missing archive")
	}
	if !strings.Contains(errOut, "unzip:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: unzip") {
		t.Errorf("help = %q", out)
	}
}
