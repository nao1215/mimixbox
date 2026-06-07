package tr_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/tr"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := tr.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"translate range", "hello\n", []string{"a-z", "A-Z"}, "HELLO\n"},
		{"translate literal", "hello", []string{"el", "ip"}, "hippo"},
		{"delete set", "hello world\n", []string{"-d", "lo"}, "he wrd\n"},
		{"delete range", "abc123\n", []string{"-d", "0-9"}, "abc\n"},
		{"squeeze", "aaabbbccc\n", []string{"-s", "a-z"}, "abc\n"},
		{"squeeze specific", "aaabbbccc", []string{"-s", "a"}, "abbbccc"},
		{"complement delete", "abc123\n", []string{"-cd", "0-9"}, "123"},
		{"complement translate", "abc123", []string{"-c", "0-9", "x"}, "xxx123"},
		{"digit class delete", "a1b2c3\n", []string{"-d", "[:digit:]"}, "abc\n"},
		{"upper class", "hello", []string{"[:lower:]", "[:upper:]"}, "HELLO"},
		{"space class squeeze", "a   b  c", []string{"-s", "[:space:]"}, "a b c"},
		{"newline escape", "a\nb\n", []string{"\\n", "_"}, "a_b_"},
		{"octal escape", "aXb", []string{"\\130", "_"}, "a_b"},
		{"translate pad shorter set2", "abcd", []string{"a-d", "x"}, "xxxx"},
		{"delete then squeeze", "aabbccdd", []string{"-ds", "a", "b"}, "bccdd"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "hello\n")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "tr: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestRunMissingSet2(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "hello\n", "a-z")
	if err == nil {
		t.Fatal("expected error for missing SET2 in translate mode")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "tr: missing operand after 'a-z'") {
		t.Errorf("stderr = %q, want missing operand after message", errOut)
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: tr") {
		t.Errorf("--help out = %q", out)
	}
}
