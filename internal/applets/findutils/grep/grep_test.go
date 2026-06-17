package grep_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/findutils/grep"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := grep.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return -1
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := grep.New()
	if got := c.Name(); got != "grep" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestStdinBasic(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "apple\nbanana\ncherry\n", "an")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "banana\n" {
		t.Errorf("out = %q, want banana", out)
	}
}

func TestFlags(t *testing.T) {
	t.Parallel()
	in := "Apple\nbanana\nAPPLE pie\n"
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"ignore-case", []string{"-i", "apple"}, "Apple\nAPPLE pie\n"},
		{"invert", []string{"-v", "banana"}, "Apple\nAPPLE pie\n"},
		{"line-number", []string{"-n", "banana"}, "2:banana\n"},
		{"count", []string{"-c", "p"}, "2\n"},
		{"word", []string{"-w", "pie"}, "APPLE pie\n"},
		{"fixed", []string{"-F", "APPLE pie"}, "APPLE pie\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, in, tt.args...)
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestMultiplePatterns(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "one\ntwo\nthree\n", "-e", "one", "-e", "three")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "one\nthree\n" {
		t.Errorf("out = %q", out)
	}
}

func TestNoMatchExit1(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello\n", "zzz")
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if code := exitCode(t, err); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
}

// errReader fails on the first read, modeling an unreadable input stream.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("input error") }

func TestStdinReadErrorExitsTwoNotSilentNoMatch(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: errReader{}, Out: out, Err: errBuf}
	// A real read failure must be reported as a grep error (exit 2), not hidden
	// behind the "no matches" exit 1 with empty stderr (issue #950).
	err := grep.New().Run(context.Background(), io, []string{"foo"})
	if code := exitCode(t, err); code != 2 {
		t.Fatalf("exit = %d, want 2 on read error", code)
	}
	if !strings.Contains(errBuf.String(), "grep:") {
		t.Errorf("stderr should report the read error, got %q", errBuf.String())
	}
}

func TestInvalidRegexExit2(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "[")
	if code := exitCode(t, err); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "grep:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestMissingPatternExit2(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "")
	if code := exitCode(t, err); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
}

func TestQuiet(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "match me\n", "-q", "match")
	if out != "" {
		t.Errorf("quiet out = %q, want empty", out)
	}
	if exitCode(t, err) != 0 {
		t.Errorf("quiet exit = %v, want 0", err)
	}
}

func TestFilesWithName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("foo\nbar\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("baz\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Multiple files -> filename prefix on matches.
	out, _, err := run(t, "", "foo", a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != a+":foo\n" {
		t.Errorf("out = %q, want %q", out, a+":foo\n")
	}

	// -l prints only file names with matches.
	out, _, err = run(t, "", "-l", "ba", a, b)
	if err != nil {
		t.Fatalf("-l err = %v", err)
	}
	if !strings.Contains(out, a) || !strings.Contains(out, b) {
		t.Errorf("-l out = %q, want both files", out)
	}
}

func TestRecursive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(sub, "deep.txt")
	if err := os.WriteFile(f, []byte("needle\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-r", "needle", dir)
	if err != nil {
		t.Fatalf("-r err = %v", err)
	}
	if !strings.Contains(out, f) {
		t.Errorf("-r out = %q, want path %q", out, f)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "x", "/no/such/file/here")
	if code := exitCode(t, err); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "grep:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: grep") {
		t.Errorf("help = %q", out)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}
}
