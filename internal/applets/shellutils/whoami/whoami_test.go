package whoami_test

import (
	"bytes"
	"context"
	"errors"
	"os/user"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/whoami"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := whoami.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRun(t *testing.T) {
	t.Parallel()
	u, err := user.Current()
	if err != nil {
		t.Fatalf("user.Current() error = %v", err)
	}
	want := u.Username + "\n"

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestRunExtraOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "bogus")
	if err == nil {
		t.Fatal("expected error")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want it to contain %q", errOut, "extra operand")
	}
	if code := exitCode(err); code == 0 {
		t.Errorf("exit code = %d, want non-zero", code)
	}
}

// exitCode extracts the process exit code an error maps to. A nil error is
// success; a *command.ExitError carries its own code; any other error is a
// generic failure.
func exitCode(err error) int {
	if err == nil {
		return command.ExitSuccess
	}
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return command.ExitFailure
}
