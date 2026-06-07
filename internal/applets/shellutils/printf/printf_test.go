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
