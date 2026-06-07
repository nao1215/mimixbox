package uniq_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/uniq"
	"github.com/nao1215/mimixbox/internal/command"
)

func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := uniq.New().Run(context.Background(), io, args)
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
		{
			name:  "basic adjacent dedup",
			stdin: "a\na\nb\nc\nc\nc\n",
			args:  nil,
			want:  "a\nb\nc\n",
		},
		{
			name:  "count prefixes occurrences",
			stdin: "a\na\nb\nc\nc\nc\n",
			args:  []string{"-c"},
			want:  "      2 a\n      1 b\n      3 c\n",
		},
		{
			name:  "repeated only prints duplicates",
			stdin: "a\na\nb\nc\nc\nc\n",
			args:  []string{"-d"},
			want:  "a\nc\n",
		},
		{
			name:  "unique only prints non-repeated",
			stdin: "a\na\nb\nc\nc\nc\n",
			args:  []string{"-u"},
			want:  "b\n",
		},
		{
			name:  "ignore case",
			stdin: "A\na\nB\n",
			args:  []string{"-i"},
			want:  "A\nB\n",
		},
		{
			name:  "long flags",
			stdin: "a\na\nb\n",
			args:  []string{"--count"},
			want:  "      2 a\n      1 b\n",
		},
		{
			name:  "non-adjacent are not merged",
			stdin: "a\nb\na\n",
			args:  nil,
			want:  "a\nb\na\n",
		},
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

func TestRunInputFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(in, []byte("x\nx\ny\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := runStdin(t, "", in)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "x\ny\n" {
		t.Errorf("out = %q, want %q", out, "x\ny\n")
	}
}

func TestRunOutputFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	in := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(in, []byte("p\np\nq\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runStdin(t, "", in, outPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty (written to OUTPUT)", stdout)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "p\nq\n" {
		t.Errorf("output file = %q, want %q", string(got), "p\nq\n")
	}
}

func TestRunMissingInput(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "/no/such/file")
	if err == nil {
		t.Fatal("expected error for missing input")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(err.Error(), "/no/such/file:") {
		t.Errorf("error = %q, want file error", err.Error())
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := runStdin(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: uniq") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = runStdin(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "uniq (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
