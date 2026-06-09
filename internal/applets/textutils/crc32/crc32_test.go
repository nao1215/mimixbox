package crc32

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

func TestCrc32Stdin(t *testing.T) {
	t.Parallel()
	// CRC-32 (IEEE) of "hello\n" is 0x363a3020.
	got, err := run(t, "hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "363a3020  -\n" {
		t.Errorf("crc32 stdin = %q, want %q", got, "363a3020  -\n")
	}
}

func TestCrc32File(t *testing.T) {
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
	if got != "363a3020  "+f+"\n" {
		t.Errorf("crc32 file = %q", got)
	}
}

func TestCrc32Missing(t *testing.T) {
	t.Parallel()
	if _, err := run(t, "", "/no/such/crc/file"); err == nil {
		t.Errorf("missing file should fail")
	}
}
