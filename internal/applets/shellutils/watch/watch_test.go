package watch

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// runBounded runs watch with a context that cancels after a short delay, so the
// refresh loop renders a few times and then returns (watch otherwise loops
// forever). It returns the collected stdout and the run error.
func runBounded(t *testing.T, timeout time.Duration, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := New().Run(ctx, io, args)
	return out.String(), errBuf.String(), err
}

func TestRendersWithoutTitle(t *testing.T) {
	t.Parallel()
	out, _, err := runBounded(t, 150*time.Millisecond, "-t", "-n", "0.05", "echo", "hello")
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
	out, _, err := runBounded(t, 150*time.Millisecond, "-n", "0.05", "echo", "hi")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Every 0.05s: echo hi") {
		t.Errorf("missing header in %q", out)
	}
}

func TestCommandError(t *testing.T) {
	t.Parallel()
	out, _, err := runBounded(t, 150*time.Millisecond, "-t", "-n", "0.05", "false")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "watch:") {
		t.Errorf("expected error text in %q", out)
	}
}

func TestCancelInterruptsHungChild(t *testing.T) {
	t.Parallel()
	// The first render runs "sleep 30"; cancelling the context must kill that
	// child (via exec.CommandContext) so Run returns promptly instead of
	// blocking for the full sleep.
	start := time.Now()
	_, _, err := runBounded(t, 200*time.Millisecond, "-t", "sleep", "30")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Run took %v; a hung child was not interrupted by cancellation", elapsed)
	}
}

func TestMissingCommand(t *testing.T) {
	t.Parallel()
	_, _, err := runBounded(t, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing command operand") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidInterval(t *testing.T) {
	t.Parallel()
	_, _, err := runBounded(t, 100*time.Millisecond, "-n", "0", "echo", "x")
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
