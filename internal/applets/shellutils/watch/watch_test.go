package watch

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func runCancelled(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	// An already-cancelled context makes Run render exactly once and return.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := New().Run(ctx, io, args)
	return out.String(), errBuf.String(), err
}

func TestRendersOnceWithoutTitle(t *testing.T) {
	t.Parallel()
	out, _, err := runCancelled(t, "-t", "echo", "hello")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "hello\n") {
		t.Errorf("out = %q", out)
	}
	if !strings.HasPrefix(out, clearScreen) {
		t.Errorf("output should start by clearing the screen, got %q", out)
	}
	if strings.Contains(out, "Every") {
		t.Errorf("title should be suppressed with -t, got %q", out)
	}
}

func TestRendersHeader(t *testing.T) {
	t.Parallel()
	out, _, err := runCancelled(t, "-n", "1", "echo", "hi")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Every 1s: echo hi") {
		t.Errorf("missing header in %q", out)
	}
}

func TestCommandError(t *testing.T) {
	t.Parallel()
	out, _, err := runCancelled(t, "-t", "false")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "watch:") {
		t.Errorf("expected error text in %q", out)
	}
}

func TestMissingCommand(t *testing.T) {
	t.Parallel()
	_, _, err := runCancelled(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing command operand") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidInterval(t *testing.T) {
	t.Parallel()
	_, _, err := runCancelled(t, "-n", "0", "echo", "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "interval must be positive") {
		t.Errorf("err = %v", err)
	}
}

func TestHeaderFormat(t *testing.T) {
	t.Parallel()
	if got := header(2.5, []string{"ls", "-l"}); got != "Every 2.5s: ls -l\n\n" {
		t.Errorf("header = %q", got)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "watch" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
