package cttyhack

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
)

func TestRunsProgram(t *testing.T) {
	// Use a repo-local fake echo so the test runs even on a stripped-down
	// host without /bin/echo. UseOnly hides host commands entirely.
	fakecmd.UseOnly(t, "echo")
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"echo", "hi"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out.String(), "hi") {
		t.Errorf("output = %q", out.String())
	}
}

func TestHelpExitStatus(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("help missing exit status section = %q", out.String())
	}
}

func TestMissingOperand(t *testing.T) {
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	err := New().Run(context.Background(), io, nil)
	if err == nil {
		t.Fatal("missing PROGRAM should fail")
	}
	var ee *command.ExitError
	if errors.As(err, &ee) && ee.Code == 0 {
		t.Errorf("expected non-zero exit, got %d", ee.Code)
	}
}
