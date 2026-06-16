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

// TestCopyOutMissingFile covers copyOut's os.Stat error branch.
func TestCopyOutMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, []byte("no_such_file_here.txt\n"), "-o")
	if err == nil {
		t.Fatal("expected error when a named file is missing")
	}
	if !strings.Contains(errOut, "cpio:") {
		t.Errorf("stderr = %q, want cpio: prefix", errOut)
	}
}

// TestCopyOutSkipsBlankNames covers the blank-name continue branch of copyOut.
func TestCopyOutSkipsBlankNames(t *testing.T) {
	// Uses t.Chdir; cannot be parallel.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("only.txt", []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Blank lines and whitespace-only lines must be ignored.
	out, _, err := run(t, []byte("\n  \nonly.txt\n\n"), "-o")
	if err != nil {
		t.Fatalf("copy-out err = %v", err)
	}
	listOut, _, err := run(t, []byte(out), "-i", "-t")
	if err != nil {
		t.Fatalf("list err = %v", err)
	}
	if got := strings.Count(strings.TrimSpace(listOut), "\n"); got != 0 {
		t.Errorf("expected exactly one listed entry, got %q", listOut)
	}
	if !strings.Contains(listOut, "only.txt") {
		t.Errorf("list = %q, want only.txt", listOut)
	}
}

// TestDirectoryRoundTrip covers copyOut for a directory (no data) and
// extractEntry's IsDir branch.
func TestDirectoryRoundTrip(t *testing.T) {
	// Uses t.Chdir; cannot be parallel.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("adir", 0o755); err != nil {
		t.Fatal(err)
	}

	archive := archiveOf(t, "adir")

	extractDir := filepath.Join(dir, "out")
	if err := os.Mkdir(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(extractDir)
	if _, errOut, err := run(t, archive, "-i"); err != nil {
		t.Fatalf("extract err = %v (stderr=%q)", err, errOut)
	}
	info, err := os.Stat(filepath.Join(extractDir, "adir"))
	if err != nil {
		t.Fatalf("stat extracted dir: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("extracted adir is not a directory")
	}
}

// TestTruncatedArchive covers copyIn's short-read error path: a header that
// promises a name longer than the remaining bytes.
func TestTruncatedArchive(t *testing.T) {
	// Uses t.Chdir; cannot be parallel.
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("f.txt", []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := archiveOf(t, "f.txt")
	// Cut off the archive mid-entry (keep the header but drop the body).
	truncated := archive[:115]

	_, errOut, err := run(t, truncated, "-i", "-t")
	if err == nil {
		t.Fatal("expected error for a truncated archive")
	}
	if !strings.Contains(errOut, "cpio:") {
		t.Errorf("stderr = %q, want cpio: prefix", errOut)
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
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
