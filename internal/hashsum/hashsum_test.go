package hashsum_test

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec // test exercises the md5sum applet
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/hashsum"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	c := hashsum.New("md5sum", "synopsis", md5.New)
	err := c.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// md5hex returns the md5 digest of s as the applet would print it.
func md5hex(s string) string {
	sum := md5.Sum([]byte(s)) //nolint:gosec // matches the applet under test
	return fmt.Sprintf("%x", sum)
}

func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := hashsum.New("md5sum", "describe me", md5.New)
	if c.Name() != "md5sum" {
		t.Errorf("Name() = %q, want md5sum", c.Name())
	}
	if c.Synopsis() != "describe me" {
		t.Errorf("Synopsis() = %q, want 'describe me'", c.Synopsis())
	}
}

func TestDigestStdin(t *testing.T) {
	t.Parallel()
	// No operands: stdin is digested and named "-" (two spaces).
	out, _, err := run(t, "abc\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := md5hex("abc\n") + "  -\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestDigestStdinDash(t *testing.T) {
	t.Parallel()
	// An explicit "-" operand is also stdin.
	out, _, err := run(t, "hello\n", "-")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := md5hex("hello\n") + "  -\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestDigestFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := md5hex("test\n") + "  " + f + "\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestDigestTwoSpaceSeparator(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "  -\n") {
		t.Errorf("expected two-space separator before '-', got %q", out)
	}
}

func TestDigestMultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("one\n"), 0o600)
	_ = os.WriteFile(b, []byte("two\n"), 0o600)
	out, _, err := run(t, "", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := md5hex("one\n") + "  " + a + "\n" + md5hex("two\n") + "  " + b + "\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "md5sum: /no/such/file: No such file or directory\n" {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, errOut, err := run(t, "", dir)
	if err == nil {
		t.Fatal("expected error for directory")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "md5sum: "+dir+": It is directory\n" {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestCheckOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(f, []byte("test\n"), 0o600)
	list := filepath.Join(dir, "sums.txt")
	_ = os.WriteFile(list, []byte(md5hex("test\n")+"  "+f+"\n"), 0o600)

	out, _, err := run(t, "", "-c", list)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != f+": OK\n" {
		t.Errorf("out = %q, want %q", out, f+": OK\n")
	}
}

func TestCheckFailed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(f, []byte("test\n"), 0o600)
	list := filepath.Join(dir, "sums.txt")
	_ = os.WriteFile(list, []byte("00000000000000000000000000000000  "+f+"\n"), 0o600)

	out, _, err := run(t, "", "-c", list)
	if err == nil {
		t.Fatal("expected failure when digest does not match")
	}
	if out != f+": FAILED\n" {
		t.Errorf("out = %q, want %q", out, f+": FAILED\n")
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: md5sum") {
		t.Errorf("--help out = %q", out)
	}
}

// TestDifferentHash confirms the constructor is honored: a sha256-backed
// command produces the sha256 digest, not md5.
func TestDifferentHash(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader("abc\n"), Out: out, Err: &bytes.Buffer{}}
	c := hashsum.New("sha256sum", "s", sha256.New)
	if err := c.Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	sum := sha256.Sum256([]byte("abc\n"))
	want := fmt.Sprintf("%x", sum) + "  -\n"
	if out.String() != want {
		t.Errorf("out = %q, want %q", out.String(), want)
	}
}
