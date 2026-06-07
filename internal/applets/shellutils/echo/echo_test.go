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
		{"help is literal", []string{"--help"}, "--help\n"},
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
