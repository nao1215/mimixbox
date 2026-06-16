package echo_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/echo"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := echo.New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"plain", []string{"hello", "world"}, "hello world\n"},
		{"no args", nil, "\n"},
		{"no newline", []string{"-n", "hi"}, "hi"},
		{"unknown flag is literal", []string{"-x", "y"}, "-x y\n"},
		{"help not first is literal", []string{"foo", "--help"}, "foo --help\n"},
		{"version not first is literal", []string{"foo", "--version"}, "foo --version\n"},
		{"escapes off by default", []string{`a\tb`}, "a\\tb\n"},
		{"escapes on", []string{"-e", `a\tb`}, "a\tb\n"},
		{"combined flags", []string{"-ne", `a\nb`}, "a\nb"},
		{"escape c stops output", []string{"-e", `ab\ccd`}, "ab"},
		{"hex escape", []string{"-e", `\x41`}, "A\n"},
		{"octal escape", []string{"-e", `\0101`}, "A\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestHelpAsFirstArg verifies that a leading --help prints usage rather than
// echoing the literal text, matching GNU's standalone echo.
func TestHelpAsFirstArg(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Usage: echo") {
		t.Errorf("help out = %q", out)
	}
}

// TestVersionAsFirstArg verifies that a leading --version prints the version
// line rather than echoing the literal text.
func TestVersionAsFirstArg(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--version")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "echo (mimixbox)") {
		t.Errorf("version out = %q", out)
	}
}

// TestHelpSections asserts `echo --help` renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := echo.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: echo", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}
