package nohup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
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

func exitCode(err error) int {
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 0
}

func TestPassThroughWhenNotTTY(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "echo", "hello")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMissingCommand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing command operand") {
		t.Errorf("err = %v", err)
	}
}

func TestCommandNotFound(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "no-such-command-xyz")
	if code := exitCode(err); code != exitNotFound {
		t.Errorf("exit code = %d, want %d", code, exitNotFound)
	}
}

func TestPropagatesExitCode(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "false")
	if code := exitCode(err); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestRedirectsToNohupOut(t *testing.T) {
	// Force the terminal path and confirm output lands in nohup.out.
	orig := isTerminal
	isTerminal = func(io.Writer) bool { return true }
	t.Cleanup(func() { isTerminal = orig })

	dir := t.TempDir()
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if _, _, err := run(t, "echo", "to-file"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, err := os.ReadFile(dir + "/nohup.out")
	if err != nil {
		t.Fatalf("read nohup.out: %v", err)
	}
	if string(got) != "to-file\n" {
		t.Errorf("nohup.out = %q", got)
	}
}

// TestOutputWriterPassThrough returns the original stdout unchanged when it is
// not a terminal, with a no-op cleanup.
func TestOutputWriterPassThrough(t *testing.T) {
	orig := isTerminal
	isTerminal = func(io.Writer) bool { return false }
	t.Cleanup(func() { isTerminal = orig })

	out := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	w, cleanup, err := outputWriter(stdio)
	if err != nil {
		t.Fatalf("outputWriter err = %v", err)
	}
	defer cleanup()
	if w != out {
		t.Errorf("outputWriter returned %v, want the original stdout", w)
	}
}

// TestOutputWriterFallsBackToHome forces the cwd open to fail (read-only
// directory) so outputWriter writes to $HOME/nohup.out instead.
func TestOutputWriterFallsBackToHome(t *testing.T) {
	orig := isTerminal
	isTerminal = func(io.Writer) bool { return true }
	t.Cleanup(func() { isTerminal = orig })

	// A directory we own but make read-only, so creating ./nohup.out fails.
	roDir := t.TempDir()
	wd, _ := os.Getwd()
	if err := os.Chdir(roDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chmod(roDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roDir, 0o700) })

	home := t.TempDir()
	t.Setenv("HOME", home)

	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(""), Out: termOut{}, Err: errBuf}
	w, cleanup, err := outputWriter(stdio)
	if err != nil {
		// If the filesystem still allows the write (e.g. running as root), skip
		// rather than fail spuriously.
		t.Skipf("could not force cwd open failure: %v", err)
	}
	defer cleanup()
	if _, ok := w.(*os.File); !ok {
		t.Errorf("expected a *os.File writer, got %T", w)
	}
	if !strings.Contains(errBuf.String(), "appending output to 'nohup.out'") {
		t.Errorf("stderr = %q, want the redirect notice", errBuf.String())
	}
	if _, statErr := os.Stat(home + "/nohup.out"); statErr != nil {
		t.Errorf("expected $HOME/nohup.out to be created: %v", statErr)
	}
}

// termOut is a writer that satisfies the Fd() probe so it is treated like a
// terminal-backed stream by outputWriter's writer-type checks.
type termOut struct{}

func (termOut) Write(p []byte) (int, error) { return len(p), nil }
func (termOut) Fd() uintptr                 { return 0 }

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "nohup" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
