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

// TestHelpSections asserts a hashsum applet's --help renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := hashsum.New("sha256sum", "synopsis", sha256.New).Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: sha256sum", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}

// TestCheckStdin verifies that, with no operand, -c reads the digest list from
// standard input rather than from a file.
func TestCheckStdin(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	list := md5hex("test\n") + "  " + f + "\n"
	out, _, err := run(t, list, "-c")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != f+": OK\n" {
		t.Errorf("out = %q, want %q", out, f+": OK\n")
	}
}

// TestCheckListMissing verifies that the digest-list file itself not existing is
// reported on stderr and is a failure.
func TestCheckListMissing(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "-c", "/no/such/list.txt")
	if err == nil {
		t.Fatal("expected error when the checksum list cannot be opened")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.HasPrefix(errOut, "md5sum: ") || !strings.Contains(errOut, "/no/such/list.txt") {
		t.Errorf("stderr = %q, want md5sum-prefixed error mentioning the list", errOut)
	}
}

// TestCheckMalformedLine verifies that a line that is not "<digest>  <file>" is
// reported as improperly formatted and turns into a failure, while a valid line
// in the same list is still verified.
func TestCheckMalformedLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	list := filepath.Join(dir, "sums.txt")
	// First line cannot be split into exactly digest+file; second is valid.
	// A blank line is also present and must be skipped silently.
	content := "garbage-with-no-separator-token\n\n" + md5hex("test\n") + "  " + f + "\n"
	if err := os.WriteFile(list, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := run(t, "", "-c", list)
	if err == nil {
		t.Fatal("expected failure when a line is improperly formatted")
	}
	if out != f+": OK\n" {
		t.Errorf("out = %q, want %q", out, f+": OK\n")
	}
	if !strings.Contains(errOut, "improperly formatted checksum line") {
		t.Errorf("stderr = %q, want improperly-formatted message", errOut)
	}
}

// TestCheckSingleSpaceSeparator verifies parseLine's fallback: a line whose
// digest and filename are separated by a single space (so strings.Fields yields
// exactly two fields) is still accepted.
func TestCheckSingleSpaceSeparator(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f, []byte("test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	list := filepath.Join(dir, "sums.txt")
	// Single space, not the canonical two-space separator.
	if err := os.WriteFile(list, []byte(md5hex("test\n")+" "+f+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-c", list)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != f+": OK\n" {
		t.Errorf("out = %q, want %q", out, f+": OK\n")
	}
}

// TestCheckListReferencesMissingFile verifies that a well-formed line naming a
// file that does not exist is reported on stderr and fails, without printing OK
// or FAILED for it.
func TestCheckListReferencesMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	list := filepath.Join(dir, "sums.txt")
	missing := filepath.Join(dir, "ghost.txt")
	if err := os.WriteFile(list, []byte(md5hex("x")+"  "+missing+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := run(t, "", "-c", list)
	if err == nil {
		t.Fatal("expected failure when a listed file cannot be opened")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "ghost.txt") {
		t.Errorf("stderr = %q, want mention of the missing file", errOut)
	}
}

// TestCheckMixedOKAndFailed verifies that within a single list a matching and a
// mismatching entry each produce the right line and that the overall result is a
// failure because of the mismatch.
func TestCheckMixedOKAndFailed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	bad := filepath.Join(dir, "bad.txt")
	if err := os.WriteFile(good, []byte("good\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("bad\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	list := filepath.Join(dir, "sums.txt")
	content := md5hex("good\n") + "  " + good + "\n" +
		"00000000000000000000000000000000  " + bad + "\n"
	if err := os.WriteFile(list, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-c", list)
	if err == nil {
		t.Fatal("expected failure because one digest mismatched")
	}
	want := good + ": OK\n" + bad + ": FAILED\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestDigestMultipleFilesOneMissing verifies that a missing operand fails but
// later valid operands are still digested (the first error is kept, processing
// continues).
func TestDigestMultipleFilesOneMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.txt")
	out, errOut, err := run(t, "", missing, good)
	if err == nil {
		t.Fatal("expected error because one operand is missing")
	}
	if out != md5hex("ok\n")+"  "+good+"\n" {
		t.Errorf("out = %q, want digest of good only", out)
	}
	if !strings.Contains(errOut, "missing.txt: No such file or directory") {
		t.Errorf("stderr = %q, want missing-file message", errOut)
	}
}
