package pager

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// run uses a bytes.Buffer for stdout, which is not a terminal, so the pager
// takes its passthrough path.
func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := NewMore().Run(context.Background(), io, args)
	return out.String(), err
}

func TestPassthroughStdin(t *testing.T) {
	t.Parallel()
	for _, c := range []*Command{NewMore(), NewLess()} {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader("a\nb\nc\n"), Out: out, Err: &bytes.Buffer{}}
		if err := c.Run(context.Background(), io, nil); err != nil {
			t.Fatalf("%s error = %v", c.Name(), err)
		}
		if out.String() != "a\nb\nc\n" {
			t.Errorf("%s passthrough = %q", c.Name(), out.String())
		}
	}
}

func TestPassthroughFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "1.txt")
	f2 := filepath.Join(dir, "2.txt")
	_ = os.WriteFile(f1, []byte("one\n"), 0o644)
	_ = os.WriteFile(f2, []byte("two\n"), 0o644)

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := NewMore().Run(context.Background(), io, []string{f1, f2}); err != nil {
		t.Fatalf("error = %v", err)
	}
	if out.String() != "one\ntwo\n" {
		t.Errorf("file passthrough = %q, want %q", out.String(), "one\ntwo\n")
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "", "/no/such/pager/file"); err == nil {
		t.Errorf("missing file should fail")
	}
}

func TestIsTerminalFalseForBuffer(t *testing.T) {
	t.Parallel()
	if isTerminal(&bytes.Buffer{}) {
		t.Errorf("a bytes.Buffer must not be reported as a terminal")
	}
}
