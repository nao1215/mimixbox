package sum

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestSumStdin(t *testing.T) {
	t.Parallel()
	// Verified against BSD sum: "hello\n" -> 36979, 1 block.
	got, err := run(t, "hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "36979     1\n" {
		t.Errorf("sum stdin = %q, want %q", got, "36979     1\n")
	}
}

func TestSumFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "h.txt")
	if err := os.WriteFile(f, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := run(t, "", f)
	if err != nil {
		t.Fatal(err)
	}
	if got != "36979     1 "+f+"\n" {
		t.Errorf("sum file = %q", got)
	}
}

func TestSumMissingFile(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "", "/no/such/sum/file"); err == nil {
		t.Errorf("missing file should fail")
	}
}
