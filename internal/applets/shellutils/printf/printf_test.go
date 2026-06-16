package printf_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/printf"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := printf.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"string with newline", []string{"%s\n", "hi"}, "hi\n"},
		{"two decimals", []string{"%d-%d\n", "3", "4"}, "3-4\n"},
		{"escape in format", []string{`a\tb`}, "a\tb"},
		{"format reuse", []string{"%s ", "a", "b", "c"}, "a b c "},
		{"percent literal", []string{"100%%\n"}, "100%\n"},
		{"missing arg is empty", []string{"[%s]"}, "[]"},
		{"missing decimal is zero", []string{"%d"}, "0"},
		{"hex conversion", []string{"%x", "255"}, "ff"},
		{"octal conversion", []string{"%o", "8"}, "10"},
		{"char conversion", []string{"%c", "abc"}, "a"},
		{"b interprets escapes", []string{"%b", `x\ny`}, "x\ny"},
		{"width padding", []string{"%5s", "ab"}, "   ab"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand, got nil")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "printf: missing operand") {
		t.Errorf("stderr = %q, want it to contain %q", errOut, "printf: missing operand")
	}
}

// TestHelpSections asserts `printf --help` renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := printf.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: printf", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}

// TestMetaOnlyAsFirstArg proves --help/--version are honored only as the first
// operand (GitHub issue #758): a later --help is an ordinary operand.
func TestMetaOnlyAsFirstArg(t *testing.T) {
	t.Parallel()
	// --version as the first operand prints the version banner.
	out, _, err := run(t, "--version")
	if err != nil {
		t.Fatalf("--version err = %v", err)
	}
	if !strings.Contains(out, "printf (mimixbox)") {
		t.Errorf("--version = %q, want version banner", out)
	}
	// A later --help is a normal operand: the format is printed and no usage
	// block appears.
	out, _, err = run(t, "foo --help\n")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(out, "Usage:") {
		t.Errorf("a non-first --help must not trigger help: %q", out)
	}
	if !strings.Contains(out, "foo --help") {
		t.Errorf("expected the format to be printed, got %q", out)
	}
}
