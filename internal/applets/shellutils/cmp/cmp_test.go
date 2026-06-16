package cmp_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/cmp"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes the cmp command with empty stdin and returns stdout, stderr and
// the process exit code (via command.Execute).
func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	code := command.Execute(context.Background(), cmp.New(), io, args)
	return out.String(), errBuf.String(), code
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestIdentical(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "hello\nworld\n")
	b := writeFile(t, dir, "b", "hello\nworld\n")

	out, errOut, code := run(t, a, b)
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if out != "" || errOut != "" {
		t.Errorf("output = %q / %q, want empty", out, errOut)
	}
}

func TestDiffer(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// First difference is the 9th byte: "second" vs "SECOND"; the difference
	// is on line 2 (one newline precedes it). Verified against system cmp.
	a := writeFile(t, dir, "a", "first\nsecond\n")
	b := writeFile(t, dir, "b", "first\nSECOND\n")

	out, _, code := run(t, a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	want := a + " " + b + " differ: byte 7, line 2\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestDifferMatchesSystemCmp cross-checks the byte/line numbers against the
// system cmp binary so the implementation can't silently drift.
func TestDifferMatchesSystemCmp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "alpha\nbravo\ncharlie\n")
	b := writeFile(t, dir, "b", "alpha\nbravX\ncharlie\n")

	sysOut, ok := systemCmp(t, a, b)
	if !ok {
		t.Skip("system cmp not available")
	}

	out, _, code := run(t, a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	// system cmp: "<a> <b> differ: byte N, line L". Compare the "byte N, line L"
	// tail, which is path-independent.
	wantTail := tail(sysOut)
	gotTail := tail(out)
	if gotTail != wantTail {
		t.Errorf("differ tail = %q, want %q (system cmp: %q)", gotTail, wantTail, sysOut)
	}
}

func TestPrefix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	short := writeFile(t, dir, "short", "abc")
	long := writeFile(t, dir, "long", "abcd")

	out, errOut, code := run(t, short, long)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "cmp: EOF on "+short) {
		t.Errorf("stderr = %q, want EOF on %q", errOut, short)
	}
}

func TestSilent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "abc\n")
	b := writeFile(t, dir, "b", "abd\n")

	out, errOut, code := run(t, "-s", a, b)
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if out != "" || errOut != "" {
		t.Errorf("output = %q / %q, want empty with -s", out, errOut)
	}

	// Identical files with -s still exit 0 silently.
	c := writeFile(t, dir, "c", "abc\n")
	out, errOut, code = run(t, "-s", a, c)
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if out != "" || errOut != "" {
		t.Errorf("output = %q / %q, want empty", out, errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "abc\n")

	out, errOut, code := run(t, a, filepath.Join(dir, "nope"))
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "cmp: ") {
		t.Errorf("stderr = %q, want cmp error prefix", errOut)
	}
}

// tail returns the substring starting at "byte " from a cmp differ line.
func tail(s string) string {
	i := strings.Index(s, "byte ")
	if i < 0 {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(s[i:])
}

// systemCmp runs the host's cmp and returns its stdout and whether it ran.
func systemCmp(t *testing.T, a, b string) (string, bool) {
	t.Helper()
	path, err := exec.LookPath("cmp")
	if err != nil {
		return "", false
	}
	var out bytes.Buffer
	c := exec.Command(path, a, b)
	c.Stdout = &out
	_ = c.Run() // exit 1 on differ is expected; ignore.
	return out.String(), true
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := cmp.New()
	if c.Name() != "cmp" {
		t.Errorf("Name() = %q, want cmp", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestVerboseListsDifferingByte covers the -l output path, which prints the
// byte offset and the octal values of the two differing bytes.
func TestVerboseListsDifferingByte(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// 'A' (0101 octal) vs 'B' (0102 octal) at the first byte.
	a := writeFile(t, dir, "a", "A")
	b := writeFile(t, dir, "b", "B")

	out, _, code := run(t, "-l", a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out != "1 101 102\n" {
		t.Errorf("-l out = %q, want %q", out, "1 101 102\n")
	}
}

// TestMissingOperand covers the no-operand error (exit 2).
func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, code := run(t)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "cmp: missing operand") {
		t.Errorf("stderr = %q, want missing-operand message", errOut)
	}
}

// TestExtraOperand covers the >2-operands error (exit 2).
func TestExtraOperand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "x")
	b := writeFile(t, dir, "b", "x")
	c := writeFile(t, dir, "c", "x")
	_, errOut, code := run(t, a, b, c)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want extra-operand message", errOut)
	}
}

// TestBothStdin covers the rejection of two '-' operands (exit 2).
func TestBothStdin(t *testing.T) {
	t.Parallel()
	_, errOut, code := run(t, "-", "-")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "at most one operand may be '-'") {
		t.Errorf("stderr = %q, want both-stdin message", errOut)
	}
}

// TestPrefixEOFOnSecond exercises the eofOn==2 branch (second file is the
// shorter prefix), which the existing prefix test does not reach.
func TestPrefixEOFOnSecond(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	long := writeFile(t, dir, "long", "abcd")
	short := writeFile(t, dir, "short", "abc")

	out, errOut, code := run(t, long, short)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "cmp: EOF on "+short) {
		t.Errorf("stderr = %q, want EOF on the second file %q", errOut, short)
	}
}

// TestBytesLimit covers -n/--bytes: a difference past the limit is not seen, so
// the comparison succeeds (exit 0); within the limit it is reported (exit 1).
func TestBytesLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "abcXdef")
	b := writeFile(t, dir, "b", "abcYdef")

	// Differ at byte 4; limiting to 3 bytes hides it.
	out, errOut, code := run(t, "-n", "3", a, b)
	if code != 0 {
		t.Errorf("exit = %d, want 0 with -n 3", code)
	}
	if out != "" || errOut != "" {
		t.Errorf("output = %q / %q, want empty", out, errOut)
	}

	// Long-form --bytes=4 reaches the difference.
	out, _, code = run(t, "--bytes=4", a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1 with --bytes=4", code)
	}
	want := a + " " + b + " differ: byte 4, line 1\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestIgnoreInitialSingle covers -i N: the first N bytes of both files are
// skipped before comparing.
func TestIgnoreInitialSingle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Bytes 1-3 differ but are skipped; the rest is identical.
	a := writeFile(t, dir, "a", "XXXcommon")
	b := writeFile(t, dir, "b", "YYYcommon")

	out, errOut, code := run(t, "-i", "3", a, b)
	if code != 0 {
		t.Errorf("exit = %d, want 0 after skipping 3 bytes; stderr=%q", code, errOut)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

// TestIgnoreInitialPair covers -i N:M: different skip counts for each file, with
// reported offsets counted from the first compared byte.
func TestIgnoreInitialPair(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Skip 1 of FILE1 and 3 of FILE2; remaining "abZ" vs "abQ" differ at byte 3.
	a := writeFile(t, dir, "a", "_abZ")
	b := writeFile(t, dir, "b", "___abQ")

	out, _, code := run(t, "-i", "1:3", a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	want := a + " " + b + " differ: byte 3, line 1\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestIgnoreInitialInvalid covers a malformed --ignore-initial value (exit 2).
func TestIgnoreInitialInvalid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "x")
	b := writeFile(t, dir, "b", "x")
	_, errOut, code := run(t, "-i", "abc", a, b)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "invalid --ignore-initial") {
		t.Errorf("stderr = %q, want invalid --ignore-initial message", errOut)
	}
}

// TestPrintBytes covers -b/--print-bytes: the differ message gains the octal
// values and rendered characters of the two differing bytes, in GNU format.
func TestPrintBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "first\nsecond\n")
	b := writeFile(t, dir, "b", "first\nSECOND\n")

	out, _, code := run(t, "-b", a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	// 's' = 0163 octal, 'S' = 0123 octal; byte 7, line 2 (verified vs GNU cmp).
	want := a + " " + b + " differ: byte 7, line 2 is 163 s 123 S\n"
	if out != want {
		t.Errorf("-b out = %q, want %q", out, want)
	}
}

// TestPrintBytesControl covers sprintc's caret notation for control bytes in the
// --print-bytes output.
func TestPrintBytesControl(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := writeFile(t, dir, "a", "a\tb")
	b := writeFile(t, dir, "b", "a\nb")

	out, _, code := run(t, "-b", a, b)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	// '\t' = 11 octal, '\n' = 12 octal (GNU pads the octal field with spaces to
	// width 3); rendered ^I and ^J. Verified against GNU cmp 3.10.
	want := a + " " + b + " differ: byte 2, line 1 is  11 ^I  12 ^J\n"
	if out != want {
		t.Errorf("-b control out = %q, want %q", out, want)
	}
}

// TestHelpSections verifies that --help renders both the Examples and the
// Exit status sections supplied through WithHelp.
func TestHelpSections(t *testing.T) {
	out, _, code := run(t, "--help")
	if code != 0 {
		t.Fatalf("--help exit code = %d, want 0", code)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q section:\n%s", want, out)
		}
	}
}
