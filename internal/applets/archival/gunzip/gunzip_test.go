package gunzip_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/gunzip"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := gunzip.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// gz returns the gzip-compressed form of data.
func gz(t *testing.T, data string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeGz(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, gz(t, data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := gunzip.New()
	if got := c.Name(); got != "gunzip" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestStdin(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, gz(t, "hello world"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "hello world" {
		t.Errorf("out = %q, want hello world", out)
	}
}

func TestStdout(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "f.gz")
	writeGz(t, src, "data here")

	out, _, err := run(t, nil, "-c", src)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "data here" {
		t.Errorf("out = %q", out)
	}
	// -c keeps the input.
	if _, statErr := os.Stat(src); statErr != nil {
		t.Errorf("-c should keep the input file: %v", statErr)
	}
}

func TestFileReplace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "doc.gz")
	writeGz(t, src, "contents")

	if _, errOut, err := run(t, nil, src); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "doc"))
	if err != nil {
		t.Fatalf("expected decompressed file: %v", err)
	}
	if string(got) != "contents" {
		t.Errorf("decompressed = %q", got)
	}
	// Without -k the .gz input is removed.
	if _, statErr := os.Stat(src); statErr == nil {
		t.Error("input .gz should be removed without -k")
	}
}

func TestKeep(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "k.gz")
	writeGz(t, src, "x")
	if _, _, err := run(t, nil, "-k", src); err != nil {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(src); statErr != nil {
		t.Errorf("-k should keep input: %v", statErr)
	}
}

func TestUnknownSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "noext")
	writeGz(t, src, "x")
	_, errOut, err := run(t, nil, src)
	if err == nil {
		t.Error("expected error for unknown suffix")
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestBadGzip(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, []byte("not gzip"))
	if err == nil {
		t.Error("expected error for invalid gzip stream")
	}
	if !strings.Contains(errOut, "gunzip:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestExistingOutputNeedsForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "e.gz")
	writeGz(t, src, "x")
	// Pre-create the output so it already exists.
	if err := os.WriteFile(filepath.Join(dir, "e"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, nil, src)
	if err == nil {
		t.Error("expected error when output exists without -f")
	}
	if !strings.Contains(errOut, "already exists") {
		t.Errorf("stderr = %q", errOut)
	}
	// With -f it overwrites.
	if _, _, err := run(t, nil, "-f", src); err != nil {
		t.Fatalf("-f err = %v", err)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: gunzip") {
		t.Errorf("help = %q", out)
	}
}
