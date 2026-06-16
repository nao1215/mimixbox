package expr_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/expr"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes the expr applet with the given args and returns the captured
// stdout, stderr and the process exit code (via command.Execute).
func run(t *testing.T, args ...string) (out, errOut string, code int) {
	t.Helper()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: outBuf, Err: errBuf}
	code = command.Execute(context.Background(), expr.New(), io, args)
	return outBuf.String(), errBuf.String(), code
}

func TestRunResult(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
		code int
	}{
		{"add", []string{"1", "+", "2"}, "3\n", 0},
		{"sub negative", []string{"5", "-", "9"}, "-4\n", 0},
		{"mul", []string{"3", "*", "4"}, "12\n", 0},
		{"div", []string{"7", "/", "2"}, "3\n", 0},
		{"mod", []string{"7", "%", "2"}, "1\n", 0},
		{"less than", []string{"2", "<", "3"}, "1\n", 0},
		{"less than false", []string{"3", "<", "2"}, "0\n", 1},
		{"le", []string{"3", "<=", "3"}, "1\n", 0},
		{"eq numeric", []string{"10", "=", "10"}, "1\n", 0},
		{"ne", []string{"10", "!=", "9"}, "1\n", 0},
		{"ge", []string{"3", ">=", "4"}, "0\n", 1},
		{"gt", []string{"5", ">", "4"}, "1\n", 0},
		{"string eq", []string{"foo", "=", "foo"}, "1\n", 0},
		{"string ne", []string{"foo", "=", "bar"}, "0\n", 1},
		{"grouping", []string{"(", "1", "+", "2", ")", "*", "3"}, "9\n", 0},
		{"precedence", []string{"1", "+", "2", "*", "3"}, "7\n", 0},
		{"length", []string{"length", "abcd"}, "4\n", 0},
		{"substr", []string{"substr", "abcdef", "2", "3"}, "bcd\n", 0},
		{"substr out of range", []string{"substr", "abc", "5", "2"}, "\n", 1},
		{"index found", []string{"index", "abcdef", "cd"}, "3\n", 0},
		{"index not found", []string{"index", "abcdef", "xy"}, "0\n", 1},
		{"or first", []string{"abc", "|", "def"}, "abc\n", 0},
		{"or second", []string{"", "|", "def"}, "def\n", 0},
		{"or both zero", []string{"0", "|", "0"}, "0\n", 1},
		{"and both true", []string{"5", "&", "3"}, "5\n", 0},
		{"and one false", []string{"0", "&", "3"}, "0\n", 1},
		{"zero exits one", []string{"0"}, "0\n", 1},
		{"empty exits one", []string{""}, "\n", 1},
		{"nonzero literal", []string{"7"}, "7\n", 0},
		{"match count", []string{"abcabc", ":", "abc"}, "3\n", 0},
		{"match group", []string{"abc", ":", `a\(b\)c`}, "b\n", 0},
		{"match keyword", []string{"match", "abcabc", "abc"}, "3\n", 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, code := run(t, tt.args...)
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
			if code != tt.code {
				t.Errorf("code = %d, want %d", code, tt.code)
			}
		})
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		code int
	}{
		{"divide by zero", []string{"5", "/", "0"}, 2},
		{"modulo by zero", []string{"5", "%", "0"}, 2},
		{"missing argument", []string{"1", "+"}, 2},
		{"trailing token", []string{"1", "2"}, 2},
		{"non integer arithmetic", []string{"a", "+", "1"}, 2},
		{"unbalanced paren", []string{"(", "1", "+", "2"}, 2},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, code := run(t, tt.args...)
			if code != tt.code {
				t.Errorf("code = %d, want %d", code, tt.code)
			}
			if !strings.HasPrefix(errOut, "expr: ") {
				t.Errorf("stderr = %q, want prefix %q", errOut, "expr: ")
			}
		})
	}
}

// TestEval exercises the eval entry point directly so the evaluator can be
// covered without going through the IO layer.
func TestEval(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"add", []string{"1", "+", "2"}, "3", false},
		{"nested grouping", []string{"(", "(", "2", "+", "3", ")", "*", "2", ")"}, "10", false},
		{"div by zero", []string{"1", "/", "0"}, "", true},
		{"length unicode", []string{"length", "あいう"}, "3", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := expr.Eval(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestHelpSections asserts `expr --help` renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if code := command.Execute(context.Background(), expr.New(), io, []string{"--help"}); code != command.ExitSuccess {
		t.Fatalf("--help exit = %d, want 0", code)
	}
	for _, want := range []string{"Usage: expr", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}
