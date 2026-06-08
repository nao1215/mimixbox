package tty

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNotATTY(t *testing.T) {
	t.Parallel()
	// A strings.Reader has no Fd, so it can never be a terminal.
	out, _, err := run(t)
	if err == nil {
		t.Fatal("expected non-zero exit when stdin is not a tty")
	}
	if out != "not a tty\n" {
		t.Errorf("out = %q", out)
	}
}

func TestSilent(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-s")
	if err == nil {
		t.Fatal("expected non-zero exit")
	}
	if out != "" {
		t.Errorf("silent output should be empty, got %q", out)
	}
}

func TestTTYNameNonFile(t *testing.T) {
	t.Parallel()
	if _, ok := ttyName(strings.NewReader("")); ok {
		t.Error("a non-file reader must not be reported as a tty")
	}
}

func TestTTYNameRegularFileIsNotTTY(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	// A regular file has an Fd but is not a terminal.
	if _, ok := ttyName(f); ok {
		t.Error("a regular file must not be reported as a tty")
	}
}

func TestRunWithRegularFileStdin(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	out := &bytes.Buffer{}
	io := command.IO{In: f, Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected non-zero exit when stdin is a regular file")
	}
	if out.String() != "not a tty\n" {
		t.Errorf("out = %q", out.String())
	}
}

func TestTTYNameResolvesViaProc(t *testing.T) {
	// Pretend the descriptor is a terminal; ttyName should then resolve the
	// device name through /proc/self/fd/N, which works for any open file.
	orig := isTerminal
	isTerminal = func(int) bool { return true }
	t.Cleanup(func() { isTerminal = orig })

	p := filepath.Join(t.TempDir(), "f")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	out := &bytes.Buffer{}
	io := command.IO{In: f, Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	resolved, _ := filepath.EvalSymlinks(p)
	got := strings.TrimRight(out.String(), "\n")
	if got != p && got != resolved {
		t.Errorf("out = %q, want %q", got, p)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "tty" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
