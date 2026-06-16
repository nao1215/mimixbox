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

func TestHelpSections(t *testing.T) {
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := gzipCmd.New()
	if c.Name() != "gzip" {
		t.Errorf("Name() = %q, want %q", c.Name(), "gzip")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestKeepFlagRetainsOriginal verifies -k leaves the source file in place after
// compression instead of removing it.
func TestKeepFlagRetainsOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(name, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, "", "-k", name); err != nil {
		t.Fatalf("compress error = %v, stderr = %q", err, errOut)
	}
	if _, err := os.Stat(name); err != nil {
		t.Errorf("original should be kept with -k: %v", err)
	}
	if _, err := os.Stat(name + ".gz"); err != nil {
		t.Errorf("compressed file missing: %v", err)
	}
}

// TestRefuseOverwriteWithoutForce checks that gzip refuses to clobber an
// existing FILE.gz unless -f is given, and that -f overrides the refusal.
func TestRefuseOverwriteWithoutForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(name, []byte("payload\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Pre-existing output blocks compression without -f.
	if err := os.WriteFile(name+".gz", []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "", "-k", name)
	if err == nil {
		t.Fatal("expected error when output exists without -f")
	}
	if !strings.Contains(errOut, "already exists") {
		t.Errorf("stderr = %q, want already-exists message", errOut)
	}

	// With -f the existing output is overwritten and the command succeeds.
	if _, errOut, err := run(t, "", "-k", "-f", name); err != nil {
		t.Fatalf("force error = %v, stderr = %q", err, errOut)
	}
}

// TestDecompressUnknownSuffix checks that decompressing a file without the .gz
// suffix is rejected.
func TestDecompressUnknownSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(name, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, "", "-d", name)
	if err == nil {
		t.Fatal("expected error for unknown suffix")
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q, want unknown-suffix message", errOut)
	}
}

// TestProcessFilesContinuesAfterError verifies a missing file does not stop
// processing of the remaining valid operands, while still setting failure.
func TestProcessFilesContinuesAfterError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, "", "-k", filepath.Join(dir, "missing.txt"), good)
	if err == nil {
		t.Fatal("expected failure because one file is missing")
	}
	if !strings.Contains(errOut, "missing.txt") {
		t.Errorf("stderr = %q, want missing file message", errOut)
	}
	// The good file was still compressed despite the earlier error.
	if _, statErr := os.Stat(good + ".gz"); statErr != nil {
		t.Errorf("good file was not compressed: %v", statErr)
	}
}

// TestDashAmongFilesStreamsStdio exercises the "-" branch inside processFiles
// (rather than the single-operand fast path) by pairing "-" with a real file.
func TestDashAmongFilesStreamsStdio(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	name := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(name, []byte("body\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := run(t, "abc\n", "-k", name, "-")
	if err != nil {
		t.Fatalf("Run error = %v, stderr = %q", err, errOut)
	}
	// stdin was compressed to stdout for the "-" operand.
	if _, gzErr := gzip.NewReader(bytes.NewReader([]byte(out))); gzErr != nil {
		t.Errorf("stdout for '-' is not gzip: %v", gzErr)
	}
	// The named file was compressed on disk.
	if _, statErr := os.Stat(name + ".gz"); statErr != nil {
		t.Errorf("named file not compressed: %v", statErr)
	}
}

// TestDashStreamsStdio verifies that a "-" operand streams stdin to stdout
// rather than touching the filesystem.
func TestDashStreamsStdio(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello stream\n", "-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// Round-trip the compressed stdout to confirm it decompresses correctly.
	gr, err := gzip.NewReader(bytes.NewReader([]byte(out)))
	if err != nil {
		t.Fatalf("output is not gzip: %v", err)
	}
	got, _ := io.ReadAll(gr)
	if string(got) != "hello stream\n" {
		t.Errorf("decompressed = %q, want %q", got, "hello stream\n")
	}
}
