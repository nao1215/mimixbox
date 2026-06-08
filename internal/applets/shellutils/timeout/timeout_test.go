package timeout

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"syscall"
	"testing"
	"time"

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

func TestCommandFinishesInTime(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "5", "echo", "hello")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello\n" {
		t.Errorf("out = %q", out)
	}
}

func TestTimesOut(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "0.1", "sleep", "5")
	if code := exitCode(err); code != exitTimedOut {
		t.Errorf("exit code = %d, want %d", code, exitTimedOut)
	}
}

func TestCommandNotFound(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "1", "no-such-command-xyz")
	if code := exitCode(err); code != exitNotFound {
		t.Errorf("exit code = %d, want %d", code, exitNotFound)
	}
}

func TestPropagatesExitCode(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "5", "false")
	if code := exitCode(err); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestMissingArgs(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "5")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing duration") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidDuration(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "abc", "true")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid time interval") {
		t.Errorf("err = %v", err)
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()
	tests := map[string]time.Duration{
		"5":    5 * time.Second,
		"2m":   2 * time.Minute,
		"1h":   time.Hour,
		"1d":   24 * time.Hour,
		"0.5s": 500 * time.Millisecond,
	}
	for in, want := range tests {
		got, err := parseDuration(in)
		if err != nil {
			t.Errorf("parseDuration(%q) error = %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseDuration(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseSignal(t *testing.T) {
	t.Parallel()
	if s, err := parseSignal("KILL"); err != nil || s != syscall.SIGKILL {
		t.Errorf("parseSignal(KILL) = %v, %v", s, err)
	}
	if s, err := parseSignal("SIGTERM"); err != nil || s != syscall.SIGTERM {
		t.Errorf("parseSignal(SIGTERM) = %v, %v", s, err)
	}
	if s, err := parseSignal("9"); err != nil || s != syscall.Signal(9) {
		t.Errorf("parseSignal(9) = %v, %v", s, err)
	}
	if _, err := parseSignal("BOGUS"); err == nil {
		t.Error("expected error for unknown signal")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "timeout" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
