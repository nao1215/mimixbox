package uncompress_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/archival/lzw"
	"github.com/nao1215/mimixbox/internal/applets/archival/uncompress"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: bytes.NewReader(stdin), Out: out, Err: errBuf}
	err := uncompress.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// zbytes returns the .Z-compressed form of s.
func zbytes(t *testing.T, s string) []byte {
	t.Helper()
	var z bytes.Buffer
	if err := lzw.Compress(strings.NewReader(s), &z); err != nil {
		t.Fatal(err)
	}
	return z.Bytes()
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := uncompress.New()
	if got := c.Name(); got != "uncompress" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestStdin(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, zbytes(t, "hello there"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "hello there" {
		t.Errorf("out = %q", out)
	}
}

func TestFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zf := filepath.Join(dir, "doc.Z")
	if err := os.WriteFile(zf, zbytes(t, "the content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, nil, zf); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "doc"))
	if err != nil {
		t.Fatalf("expected decompressed file: %v", err)
	}
	if string(got) != "the content" {
		t.Errorf("decompressed = %q", got)
	}
	if _, statErr := os.Stat(zf); statErr == nil {
		t.Error(".Z input should be removed without -k")
	}
}

func TestStdoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zf := filepath.Join(dir, "s.Z")
	if err := os.WriteFile(zf, zbytes(t, "stdout please"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, nil, "-c", zf)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "stdout please" {
		t.Errorf("out = %q", out)
	}
	if _, statErr := os.Stat(zf); statErr != nil {
		t.Error("-c should keep the input")
	}
}

func TestUnknownSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "noext")
	if err := os.WriteFile(f, zbytes(t, "x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, nil, f)
	if err == nil {
		t.Error("expected error for unknown suffix")
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestBadStream(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, []byte("not a .Z stream"))
	if err == nil {
		t.Error("expected error for invalid stream")
	}
	if !strings.Contains(errOut, "uncompress:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, nil, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: uncompress") {
		t.Errorf("help = %q", out)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
