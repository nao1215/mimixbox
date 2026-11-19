package unexpand_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/unexpand"
	"github.com/nao1215/mimixbox/internal/command"
)

func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := unexpand.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunStdin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"leading 8 spaces to tab", "        x\n", nil, "\tx\n"},
		{"leading 9 spaces", "         x\n", nil, "\t x\n"},
		{"leading 10 spaces", "          x\n", nil, "\t  x\n"},
		{"non-leading kept by default", "a    b\n", nil, "a    b\n"},
		{"dash is stdin", "        x\n", []string{"-"}, "\tx\n"},
		{"all converts inner run", "a        b\n", []string{"-a"}, "a\t b\n"},
		{"all long flag", "a        b\n", []string{"--all"}, "a\t b\n"},
		{"tabs N implies all", "a    b\n", []string{"-t", "4"}, "a\t b\n"},
		{"tabs long flag", "    x\n", []string{"--tabs=4"}, "\tx\n"},
		{"plain no blanks", "hello\nworld\n", nil, "hello\nworld\n"},
		{"no trailing newline", "        x", nil, "\tx"},
		// GNU unexpand leaves a lone space that reaches a tab stop as a
		// space: a tab would not be shorter. Two blanks, or a space that the
		// run goes on past, become tabs.
		{"lone space before a tab stop kept", "aaaaaaa b\n", []string{"-a"}, "aaaaaaa b\n"},
		{"lone space at the end of a line kept", "aaaaaaa \n", []string{"-a"}, "aaaaaaa \n"},
		{"two spaces across a tab stop", "aaaaaaa  b\n", []string{"-a"}, "aaaaaaa\t b\n"},
		{"space and a tab", "aaaaaaa \tb\n", []string{"-a"}, "aaaaaaa\t\tb\n"},
		{"lone space at the start of a line converted", " b\n", []string{"-t", "1"}, "\tb\n"},
		{"lone space after a word kept with -t 1", "a b\n", []string{"-t", "1"}, "a b\n"},
		{"a tab one column wide stays a tab", "aaaaaaa\tb\n", []string{"-a"}, "aaaaaaa\tb\n"},
		{"run past two tab stops", "aaaaaaa         b\n", []string{"-a"}, "aaaaaaa\t\tb\n"},
		// Without -a only the leading blanks are rewritten; the blanks after
		// the first non-blank are written as they were, tabs included.
		{"inner tab kept by default", "a\tb\n", nil, "a\tb\n"},
		{"inner spaces and tab kept by default", "a  \t b\n", nil, "a  \t b\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, _, err := runStdin(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	out, errOut, err := runStdin(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "unexpand: /no/such/file:") {
		t.Errorf("stderr = %q, want unexpand error prefix", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: unexpand") {
		t.Errorf("--help out = %q", out)
	}
	if !strings.Contains(out, "Examples:") || !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out)
	}

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "unexpand (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
