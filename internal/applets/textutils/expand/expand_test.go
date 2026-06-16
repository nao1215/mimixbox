package expand_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/expand"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := expand.New().Run(context.Background(), io, args)
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
		{"default tab 8", "a\tb\tc\n", nil, "a       b       c\n"},
		{"dash is stdin", "a\tb\n", []string{"-"}, "a       b\n"},
		{"short tab 4", "a\tb\n", []string{"-t", "4"}, "a   b\n"},
		{"long tab", "a\tb\n", []string{"--tabs=4"}, "a   b\n"},
		{"column aware", "ab\tc\n", nil, "ab      c\n"},
		{"initial only", "  \tx\ty\n", []string{"-i"}, "        x\ty\n"},
		{"initial leading tab", "\tx\ty\n", []string{"-i"}, "        x\ty\n"},
		{"no tabs", "hello\n", nil, "hello\n"},
		{"multibyte rune column", "ab\tc\n", []string{"-t", "4"}, "ab  c\n"},
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

func TestRunMultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("a\tb\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("c\td\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "a       b\nc       d\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "expand: /no/such/file:") {
		t.Errorf("stderr = %q, want expand error prefix", errOut)
	}
}

func TestRunMissingFileContinues(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("a\tb\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, "", "/no/such/file", good)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if out != "a       b\n" {
		t.Errorf("out = %q, want good file output", out)
	}
	if !strings.Contains(errOut, "expand: /no/such/file:") {
		t.Errorf("stderr = %q, want expand error prefix", errOut)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
