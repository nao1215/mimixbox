package wc_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/wc"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := wc.New().Run(context.Background(), io, args)
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
		// Multi-column stdin uses GNU's width of 7; a single column is not padded.
		{"default", "a\nb\nc\n", nil, "      3       3       6\n"},
		{"lines only", "a\nb\nc\n", []string{"-l"}, "3\n"},
		{"words only", "foo bar baz\n", []string{"-w"}, "3\n"},
		{"bytes only", "hello", []string{"-c"}, "5\n"},
		{"chars multibyte", "あ\n", []string{"-m"}, "2\n"},
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

func TestRunFileWidth(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "w.txt")
	_ = os.WriteFile(f, []byte("a\nb\nc\n"), 0o600)

	// A named file is padded to the digit-width of the largest count (here 1).
	out, _, err := run(t, "", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "3 3 6 " + f + "\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunMultipleFilesTotal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("a\nb\nc\n"), 0o600)
	_ = os.WriteFile(b, []byte("x\n"), 0o600)

	out, _, err := run(t, "", "-l", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "3 " + a + "\n1 " + b + "\n4 total\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "wc: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}
