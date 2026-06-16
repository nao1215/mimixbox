package testcmd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/test"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes the test applet with args and returns its exit code plus the
// bytes written to stderr.
func run(t *testing.T, args ...string) (int, string) {
	t.Helper()
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	code := command.Execute(context.Background(), testcmd.New(), io, args)
	return code, errBuf.String()
}

func TestRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existing := filepath.Join(dir, "file")
	if err := os.WriteFile(existing, []byte("data"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	missing := filepath.Join(dir, "nope")

	tests := []struct {
		name string
		args []string
		want int
	}{
		{"-z empty true", []string{"-z", ""}, 0},
		{"-z nonempty false", []string{"-z", "x"}, 1},
		{"-n nonempty true", []string{"-n", "x"}, 0},
		{"-n empty false", []string{"-n", ""}, 1},
		{"string equal true", []string{"a", "=", "a"}, 0},
		{"string equal false", []string{"a", "=", "b"}, 1},
		{"string not-equal true", []string{"a", "!=", "b"}, 0},
		{"int eq true", []string{"1", "-eq", "1"}, 0},
		{"int eq false", []string{"1", "-eq", "2"}, 1},
		{"int gt true", []string{"2", "-gt", "1"}, 0},
		{"int gt false", []string{"1", "-gt", "2"}, 1},
		{"int ne true", []string{"1", "-ne", "2"}, 0},
		{"int le true", []string{"1", "-le", "1"}, 0},
		{"negate true expr", []string{"!", "1", "-eq", "1"}, 1},
		{"negate false expr", []string{"!", "1", "-eq", "2"}, 0},
		{"and true", []string{"1", "-eq", "1", "-a", "2", "-eq", "2"}, 0},
		{"and false", []string{"1", "-eq", "1", "-a", "2", "-eq", "3"}, 1},
		{"or true", []string{"1", "-eq", "2", "-o", "2", "-eq", "2"}, 0},
		{"or false", []string{"1", "-eq", "2", "-o", "2", "-eq", "3"}, 1},
		{"parens", []string{"(", "1", "-eq", "1", ")"}, 0},
		{"-f existing true", []string{"-f", existing}, 0},
		{"-f missing false", []string{"-f", missing}, 1},
		{"-e existing true", []string{"-e", existing}, 0},
		{"-d dir true", []string{"-d", dir}, 0},
		{"-d file false", []string{"-d", existing}, 1},
		{"-s nonempty true", []string{"-s", existing}, 0},
		{"bare nonempty true", []string{"hello"}, 0},
		{"bare empty false", []string{""}, 1},
		{"no args false", nil, 1},
		{"malformed missing operand", []string{"1", "-eq"}, 2},
		{"malformed unclosed paren", []string{"(", "1", "-eq", "1"}, 2},
		{"malformed non-integer", []string{"a", "-eq", "1"}, 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			code, _ := run(t, tt.args...)
			if code != tt.want {
				t.Errorf("exit code = %d, want %d", code, tt.want)
			}
		})
	}
}

func TestMalformedMessage(t *testing.T) {
	t.Parallel()
	code, stderr := run(t, "1", "-eq")
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.HasPrefix(stderr, "test: ") {
		t.Errorf("stderr = %q, want prefix %q", stderr, "test: ")
	}
}

// TestHelpSections asserts `test --help` (sole argument) renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := testcmd.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: test", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}

// TestMetaOnlyWhenSoleArg proves --help/--version are honored only as the sole
// argument (GitHub issue #759): `test foo --help` evaluates as an expression.
func TestMetaOnlyWhenSoleArg(t *testing.T) {
	t.Parallel()
	capture := func(args ...string) (string, int) {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		code := command.Execute(context.Background(), testcmd.New(), io, args)
		return out.String(), code
	}
	// Sole --help / --version produce metadata and exit 0.
	if out, code := capture("--help"); code != 0 || !strings.Contains(out, "Usage: test") {
		t.Errorf("test --help: code=%d out=%q", code, out)
	}
	if out, code := capture("--version"); code != 0 || !strings.Contains(out, "test (mimixbox)") {
		t.Errorf("test --version: code=%d out=%q", code, out)
	}
	// A non-sole --help is an ordinary operand: it is evaluated, not shown as help.
	if out, _ := capture("foo", "--help"); strings.Contains(out, "Usage: test") {
		t.Errorf("test foo --help must not print help: %q", out)
	}
}
