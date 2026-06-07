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

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "unexpand (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
