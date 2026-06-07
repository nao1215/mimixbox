package gzipCmd_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gzipCmd "github.com/nao1215/mimixbox/internal/applets/archival/gzip"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := gzipCmd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// TestRoundTripFile compresses a file in place then decompresses it and asserts
// the content matches the original.
func TestRoundTripFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "data.txt")
	content := []byte("the quick brown fox jumps over the lazy dog\n")
	if err := os.WriteFile(name, content, 0o600); err != nil {
		t.Fatal(err)
	}

	// Compress in place: data.txt -> data.txt.gz, original removed.
	if _, errOut, err := run(t, "", name); err != nil {
		t.Fatalf("compress error = %v, stderr = %q", err, errOut)
	}
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Errorf("original %s should have been removed", name)
	}
	gz := name + ".gz"
	if _, err := os.Stat(gz); err != nil {
		t.Fatalf("compressed file %s missing: %v", gz, err)
	}

	// Decompress: data.txt.gz -> data.txt, .gz removed.
	if _, errOut, err := run(t, "", "-d", gz); err != nil {
		t.Fatalf("decompress error = %v, stderr = %q", err, errOut)
	}
	if _, err := os.Stat(gz); !os.IsNotExist(err) {
		t.Errorf("compressed %s should have been removed", gz)
	}
	got, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("round-trip content = %q, want %q", got, content)
	}
}

// TestStdoutKeepsInput verifies -c writes compressed bytes to stdout and leaves
// the input file untouched.
func TestStdoutKeepsInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "data.txt")
	content := []byte("hello stdout\n")
	if err := os.WriteFile(name, content, 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, "", "-c", name)
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}

	// Input file must still exist with original content.
	if got, rerr := os.ReadFile(name); rerr != nil || !bytes.Equal(got, content) {
		t.Errorf("input file changed: got %q err %v", got, rerr)
	}
	if _, serr := os.Stat(name + ".gz"); !os.IsNotExist(serr) {
		t.Errorf("-c should not create %s.gz", name)
	}

	// Decompress the stdout bytes with the stdlib and compare.
	gr, err := gzip.NewReader(strings.NewReader(out))
	if err != nil {
		t.Fatalf("stdout is not valid gzip: %v", err)
	}
	got, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("decompressed stdout = %q, want %q", got, content)
	}
}

// TestStdinStdout compresses standard input to standard output, then verifies
// the result both with the stdlib and by feeding it back through -d.
func TestStdinStdout(t *testing.T) {
	t.Parallel()
	content := "stream me through gzip\n"

	out, errOut, err := run(t, content)
	if err != nil {
		t.Fatalf("compress stdin error = %v, stderr = %q", err, errOut)
	}

	// Verify with the stdlib directly.
	gr, err := gzip.NewReader(strings.NewReader(out))
	if err != nil {
		t.Fatalf("stdout is not valid gzip: %v", err)
	}
	got, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != content {
		t.Errorf("stdlib decompress = %q, want %q", got, content)
	}

	// Verify by decompressing through the applet (stdin -> stdout).
	back, errOut, err := run(t, out, "-d")
	if err != nil {
		t.Fatalf("decompress stdin error = %v, stderr = %q", err, errOut)
	}
	if back != content {
		t.Errorf("applet round-trip = %q, want %q", back, content)
	}
}

// TestMissingFile covers a missing-file error.
func TestMissingFile(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "gzip: /no/such/file:") {
		t.Errorf("stderr = %q, want gzip error prefix", errOut)
	}
}
