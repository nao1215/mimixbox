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
