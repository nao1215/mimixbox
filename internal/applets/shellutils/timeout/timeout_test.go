package timeout

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
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

// processAlive reports whether pid is still a live process. Signal 0 performs
// only the existence/permission check without delivering a signal.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func TestKillsDescendantProcesses(t *testing.T) {
	// A wrapped script backgrounds a child, records its PID, then sleeps. When
	// the timeout fires the whole process group must be terminated, so the
	// backgrounded descendant must not be left running (issue #951).
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	script := filepath.Join(dir, "child.sh")
	content := "#!/bin/sh\nsleep 30 &\necho $! > \"" + pidFile + "\"\nsleep 30\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	// Back the child's stdio with real files (not in-memory buffers). With a
	// pipe-backed buffer, exec's copy goroutines keep cmd.Wait() blocked until
	// the inherited pipe write end is closed by the descendant, which would
	// mask the leak; a real file makes Wait return as soon as the direct child
	// is signalled, exactly as a terminal does in the bug report.
	devnull, derr := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if derr != nil {
		t.Fatal(derr)
	}
	defer func() { _ = devnull.Close() }()
	io := command.IO{In: devnull, Out: devnull, Err: devnull}
	err := New().Run(context.Background(), io, []string{"0.3", "sh", script})
	if code := exitCode(err); code != exitTimedOut {
		t.Fatalf("exit code = %d, want %d", code, exitTimedOut)
	}

	data, rerr := os.ReadFile(pidFile)
	if rerr != nil {
		t.Fatalf("read pid file: %v", rerr)
	}
	pid, perr := strconv.Atoi(strings.TrimSpace(string(data)))
	if perr != nil || pid <= 0 {
		t.Fatalf("bad pid %q: %v", data, perr)
	}

	// The kill is asynchronous; poll briefly for the descendant to disappear.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("descendant pid %d still alive after timeout (process group not killed)", pid)
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

func TestKillAfter(t *testing.T) {
	t.Parallel()
	// A process that ignores SIGTERM is still killed after -k grace, so timeout
	// returns 124 rather than hanging. The shell busy-waits (no child process)
	// so SIGKILL on the shell itself ends it promptly.
	start := time.Now()
	_, _, err := run(t, "-s", "TERM", "-k", "0.3", "0.1", "sh", "-c", "trap '' TERM; while :; do :; done")
	if code := exitCode(err); code != exitTimedOut {
		t.Errorf("exit code = %d, want %d", code, exitTimedOut)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("took %v; -k did not force a kill", elapsed)
	}
}

func TestInvalidKillAfter(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-k", "abc", "1", "true")
	if err == nil {
		t.Fatal("expected error for invalid -k interval")
	}
	if !strings.Contains(err.Error(), "invalid time interval") {
		t.Errorf("err = %v", err)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q section:\n%s", want, out.String())
		}
	}
}
