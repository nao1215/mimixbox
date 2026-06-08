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
