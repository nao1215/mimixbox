package cat_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/cat"
	"github.com/nao1215/mimixbox/internal/command"
)

func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := cat.New().Run(context.Background(), io, args)
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
		{"plain", "hello\nworld\n", nil, "hello\nworld\n"},
		{"dash is stdin", "hi\n", []string{"-"}, "hi\n"},
		{"number all", "a\n\nb\n", []string{"-n"}, "     1\ta\n     2\t\n     3\tb\n"},
		{"number nonblank", "a\n\nb\n", []string{"-b"}, "     1\ta\n\n     2\tb\n"},
		{"show ends", "a\nb\n", []string{"-E"}, "a$\nb$\n"},
		{"show tabs", "a\tb\n", []string{"-T"}, "a^Ib\n"},
		{"squeeze blanks", "a\n\n\n\nb\n", []string{"-s"}, "a\n\nb\n"},
		{"long number flag", "x\n", []string{"--number"}, "     1\tx\n"},
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

func TestRunFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("one\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("two\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := runStdin(t, "", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "one\ntwo\n" {
		t.Errorf("out = %q, want %q", out, "one\ntwo\n")
	}
}

func TestRunNumberContinuesAcrossFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("one\n"), 0o600)
	_ = os.WriteFile(b, []byte("two\n"), 0o600)

	out, _, err := runStdin(t, "", "-n", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "     1\tone\n     2\ttwo\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
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
	if !strings.Contains(errOut, "cat: /no/such/file:") {
		t.Errorf("stderr = %q, want cat error prefix", errOut)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: cat") {
		t.Errorf("--help out = %q", out)
	}
	if !strings.Contains(out, "Examples:") {
		t.Errorf("--help missing Examples: %q", out)
	}
	if !strings.Contains(out, "Exit status:") {
		t.Errorf("--help missing Exit status: %q", out)
	}

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "cat (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
