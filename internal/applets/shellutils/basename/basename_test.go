package basename_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/basename"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := basename.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"simple", []string{"/usr/lib"}, "lib\n"},
		{"trailing slash", []string{"/usr/"}, "usr\n"},
		{"root", []string{"/"}, "/\n"},
		{"no slash", []string{"file.txt"}, "file.txt\n"},
		{"with suffix", []string{"/a/file.txt", ".txt"}, "file\n"},
		{"suffix equals name kept", []string{"/a/.txt", ".txt"}, ".txt\n"},
		{"multiple", []string{"-a", "/a/b", "/c/d"}, "b\nd\n"},
		{"suffix flag implies multiple", []string{"-s", ".go", "/x/a.go", "/y/b.go"}, "a\nb\n"},
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

func TestRunZero(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-z", "/a/b")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "b\x00" {
		t.Errorf("out = %q, want %q", out, "b\x00")
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunExtraOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "a", "b", "c")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("--help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out)
	}
}
