package mkfifo_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkfifo"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := mkfifo.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNew(t *testing.T) {
	t.Parallel()
	c := mkfifo.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Name() != "mkfifo" {
		t.Errorf("Name() = %q, want %q", c.Name(), "mkfifo")
	}
	if c.Synopsis() != "Make FIFO (named pipe)" {
		t.Errorf("Synopsis() = %q, want %q", c.Synopsis(), "Make FIFO (named pipe)")
	}
}

func TestRunCreatesFifo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "fifo")

	_, errOut, err := run(t, path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("Stat(%q) = %v", path, statErr)
	}
	if info.Mode()&os.ModeNamedPipe == 0 {
		t.Errorf("mode = %v, want a named pipe", info.Mode())
	}
}

func TestRunExistingPathErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "exists")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, path)
	if err == nil {
		t.Fatal("expected error for existing path")
	}
	if !strings.Contains(errOut, "mkfifo: can't make "+path+": already exist") {
		t.Errorf("stderr = %q, want already-exist message", errOut)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "mkfifo: missing operand") {
		t.Errorf("stderr = %q, want missing-operand message", errOut)
	}
}
