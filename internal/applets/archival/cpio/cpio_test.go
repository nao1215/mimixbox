package cpio_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/cpio"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := cpio.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := cpio.New()
	if got := c.Name(); got != "cpio" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

// archiveOf builds a newc archive from the named files by driving copy-out, and
// returns the raw archive bytes.
func archiveOf(t *testing.T, names ...string) []byte {
	t.Helper()
	out, errOut, err := run(t, []byte(strings.Join(names, "\n")+"\n"), "-o")
	if err != nil {
		t.Fatalf("copy-out err = %v (stderr=%q)", err, errOut)
	}
	return []byte(out)
}

func TestRoundTrip(t *testing.T) {
	// Uses t.Chdir for extraction; cannot be parallel.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("a.txt", []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("sub", "b.txt"), []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive := archiveOf(t, "a.txt", "sub/b.txt")

	// List.
	out, _, err := run(t, archive, "-i", "-t")
	if err != nil {
		t.Fatalf("list err = %v", err)
	}
	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "sub/b.txt") {
		t.Errorf("list = %q", out)
	}

	// Extract into a fresh directory.
	extractDir := filepath.Join(dir, "out")
	if err := os.Mkdir(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(extractDir)
	if _, errOut, err := run(t, archive, "-i"); err != nil {
		t.Fatalf("extract err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "a.txt"))
	if err != nil {
		t.Fatalf("read a.txt: %v", err)
	}
	if string(got) != "alpha" {
		t.Errorf("a.txt = %q", got)
	}
	got, err = os.ReadFile(filepath.Join(extractDir, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("read sub/b.txt: %v", err)
	}
	if string(got) != "beta" {
		t.Errorf("sub/b.txt = %q", got)
	}
}

func TestVerbose(t *testing.T) {
	// Uses t.Chdir; cannot be parallel.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("v.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, []byte("v.txt\n"), "-o", "-v")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(errOut, "v.txt") {
		t.Errorf("verbose stderr = %q", errOut)
	}
}

func TestNoModeError(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, "-v")
	if err == nil {
		t.Error("expected error when no mode is given")
	}
	if !strings.Contains(errOut, "one of -o, -i or -t") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestUnsupportedFormat(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, "-o", "-H", "odc")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
	if !strings.Contains(errOut, "unsupported format") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestBadArchive(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, []byte("this is not cpio at all!!!"), "-i", "-t")
	if err == nil {
		t.Error("expected error for invalid archive")
	}
	if !strings.Contains(errOut, "cpio:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: cpio") {
		t.Errorf("help = %q", out)
	}
}
