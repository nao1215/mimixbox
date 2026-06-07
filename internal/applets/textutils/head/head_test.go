package head_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/head"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := head.New().Run(context.Background(), io, args)
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
		{"default 10", "1\n2\n3\n", nil, "1\n2\n3\n"},
		{"lines flag", "1\n2\n3\n4\n", []string{"-n", "2"}, "1\n2\n"},
		{"bytes flag", "hello world", []string{"-c", "5"}, "hello"},
		{"long lines flag", "1\n2\n3\n", []string{"--lines", "1"}, "1\n"},
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

func TestRunMultipleFilesHaveHeaders(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("aaa\n"), 0o600)
	_ = os.WriteFile(b, []byte("bbb\n"), 0o600)

	out, _, err := run(t, "", "-n", "1", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "==> " + a + " <==\naaa\n\n==> " + b + " <==\nbbb\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunQuietSuppressesHeaders(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(a, []byte("aaa\n"), 0o600)
	_ = os.WriteFile(b, []byte("bbb\n"), 0o600)

	out, _, err := run(t, "", "-q", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "aaa\nbbb\n" {
		t.Errorf("out = %q, want %q", out, "aaa\nbbb\n")
	}
}

func TestRunMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "head: /no/such/file:") {
		t.Errorf("stderr = %q", errOut)
	}
}
