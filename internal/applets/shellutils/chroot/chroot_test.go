package chroot_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/chroot"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := chroot.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNew(t *testing.T) {
	t.Parallel()
	if chroot.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := chroot.New()
	if got := c.Name(); got != "chroot" {
		t.Errorf("Name() = %q, want %q", got, "chroot")
	}
	want := "Run command or interactive shell with special root directory"
	if got := c.Synopsis(); got != want {
		t.Errorf("Synopsis() = %q, want %q", got, want)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if got := exitCode(err); got != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", got, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "chroot: missing operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

// TestNonRootFails verifies that, in the (non-root) test environment, running
// chroot against a directory yields a failure error without panicking and
// without writing to stdout. We do not require root.
func TestNonRootFails(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "/tmp")
	if err == nil {
		t.Fatal("expected failure when not root")
	}
	if got := exitCode(err); got != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", got, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	if !strings.Contains(errOut, "chroot: cannot change root directory to '/tmp':") {
		t.Errorf("stderr = %q, want chroot error prefix", errOut)
	}
}

// exitCode extracts the exit status the runner would use for err. Both the
// silent failure returned on usage/permission errors and an *ExitError report
// ExitFailure (1).
func exitCode(err error) int {
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return command.ExitFailure
}
