package zip_test

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	zipcmd "github.com/nao1215/mimixbox/internal/applets/archival/zip"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := zipcmd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// names returns the sorted entry names in a zip archive.
func names(t *testing.T, archive string) map[string]string {
	t.Helper()
	zr, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	m := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		buf, err := func() ([]byte, error) {
			defer func() { _ = rc.Close() }()
			return io.ReadAll(rc)
		}()
		if err != nil {
			t.Fatal(err)
		}
		m[f.Name] = string(buf)
	}
	return m
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := zipcmd.New()
	if got := c.Name(); got != "zip" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestZipFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(a, []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")

	if _, errOut, err := run(t, archive, a); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := names(t, archive)
	if len(got) != 1 {
		t.Fatalf("entries = %v, want 1", got)
	}
	for _, v := range got {
		if v != "alpha" {
			t.Errorf("content = %q, want alpha", v)
		}
	}
}

func TestZipRecurse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "d")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "x.txt"), []byte("ex"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")

	if _, errOut, err := run(t, "-r", archive, sub); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := names(t, archive)
	if len(got) != 1 {
		t.Errorf("entries = %v, want 1 file from recursion", got)
	}
}

func TestZipDirWithoutRecurseSkips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "d")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")
	_, errOut, err := run(t, archive, sub)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(errOut, "use -r to recurse") {
		t.Errorf("stderr = %q, want recurse hint", errOut)
	}
}

func TestUsageError(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "only-archive")
	if err == nil {
		t.Error("expected usage error with too few operands")
	}
	if !strings.Contains(errOut, "usage") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "o.zip"), filepath.Join(dir, "nope"))
	if err == nil {
		t.Error("expected error for missing input file")
	}
	if !strings.Contains(errOut, "zip:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: zip") {
		t.Errorf("help = %q", out)
	}
}
