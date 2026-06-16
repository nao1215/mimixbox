package kill

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// run executes the kill command in memory and returns stdout, stderr and the
// returned error.
func run(args ...string) (string, string, error) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestResolveSignal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		spec    string
		want    int
		wantErr bool
	}{
		{"short name KILL", "KILL", 9, false},
		{"lower-case kill", "kill", 9, false},
		{"full name SIGKILL", "SIGKILL", 9, false},
		{"number 9", "9", 9, false},
		{"SIGTERM", "SIGTERM", 15, false},
		{"short TERM", "TERM", 15, false},
		{"number 15", "15", 15, false},
		{"number 1", "1", 1, false},
		{"invalid name", "NOPE", 0, true},
		{"invalid number", "999", 0, true},
		{"empty", "", 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveSignal(tt.spec)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveSignal(%q) expected error, got %d", tt.spec, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveSignal(%q) unexpected error: %v", tt.spec, err)
			}
			if got != tt.want {
				t.Errorf("resolveSignal(%q) = %d, want %d", tt.spec, got, tt.want)
			}
		})
	}
}

func TestIsSignalSpec(t *testing.T) {
	t.Parallel()
	yes := []string{"9", "KILL", "SIGKILL", "15", "TERM"}
	for _, s := range yes {
		if !isSignalSpec(s) {
			t.Errorf("isSignalSpec(%q) = false, want true", s)
		}
	}
	no := []string{"foo", "999", ""}
	for _, s := range no {
		if isSignalSpec(s) {
			t.Errorf("isSignalSpec(%q) = true, want false", s)
		}
	}
}

func TestListSignals(t *testing.T) {
	t.Parallel()
	out, errBuf, err := run("-l")
	if err != nil {
		t.Fatalf("kill -l error = %v", err)
	}
	if errBuf != "" {
		t.Errorf("kill -l stderr = %q, want empty", errBuf)
	}
	if !strings.Contains(out, "SIGKILL") || !strings.Contains(out, "SIGTERM") {
		t.Errorf("kill -l out = %q, want signal names", out)
	}
	// One line per signal in the table.
	if lines := strings.Count(out, "\n"); lines != len(signals) {
		t.Errorf("kill -l printed %d lines, want %d", lines, len(signals))
	}
}

func TestInvalidPID(t *testing.T) {
	t.Parallel()
	out, errBuf, err := run("notapid")
	if err == nil {
		t.Fatal("expected error for invalid pid")
	}
	if exit := command.Execute(context.Background(), New(), command.IO{In: &bytes.Buffer{}, Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}, []string{"notapid"}); exit != command.ExitFailure {
		t.Errorf("exit = %d, want %d", exit, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("stdout = %q, want empty", out)
	}
	want := "kill: notapid: arguments must be process or job IDs"
	if !strings.Contains(errBuf, want) {
		t.Errorf("stderr = %q, want to contain %q", errBuf, want)
	}
}

func TestInvalidSignal(t *testing.T) {
	t.Parallel()
	_, errBuf, err := run("-s", "BOGUS", "1")
	if err == nil {
		t.Fatal("expected error for invalid signal")
	}
	want := "kill: BOGUS: invalid signal specification"
	if !strings.Contains(errBuf, want) {
		t.Errorf("stderr = %q, want to contain %q", errBuf, want)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errBuf, err := run()
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errBuf, "Usage: kill") {
		t.Errorf("stderr = %q, want usage", errBuf)
	}
}

// TestSignalZeroSelf sends signal 0 to the test process itself, which performs
// the existence check without actually delivering a signal. This exercises the
// full send path safely.
func TestSignalZeroSelf(t *testing.T) {
	t.Parallel()
	pid := strconv.Itoa(os.Getpid())
	_, errBuf, err := run("-s", "0", pid)
	if err != nil {
		t.Fatalf("kill -s 0 self error = %v (stderr=%q)", err, errBuf)
	}
}

// TestKillRealProcess starts a harmless long-running sleep process and kills it
// with SIGKILL via the applet, then verifies it terminated.
func TestKillRealProcess(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep not available")
	}
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	pid := strconv.Itoa(cmd.Process.Pid)

	_, errBuf, err := run("-9", pid)
	if err != nil {
		t.Fatalf("kill -9 %s error = %v (stderr=%q)", pid, err, errBuf)
	}

	if waitErr := cmd.Wait(); waitErr == nil {
		t.Fatal("expected sleep to be killed, but it exited cleanly")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "kill" {
		t.Errorf("Name() = %q, want %q", c.Name(), "kill")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestSignalNonexistentProcess exercises sendSignals' delivery-error branch by
// sending signal 0 to a PID that does not exist, which yields ESRCH.
func TestSignalNonexistentProcess(t *testing.T) {
	t.Parallel()
	// PID 2^30 is far above any real process; signal 0 only probes existence.
	deadPID := strconv.Itoa(1 << 30)
	_, errBuf, err := run("-s", "0", deadPID)
	if err == nil {
		t.Fatal("expected error signaling a non-existent process")
	}
	if !strings.Contains(errBuf, "kill: "+deadPID+":") {
		t.Errorf("stderr = %q, want delivery error for pid %s", errBuf, deadPID)
	}
}

// TestSignalContinuesAfterBadPID confirms that an invalid PID in the middle of
// the list does not stop delivery to the valid PIDs, while failure is reported.
func TestSignalContinuesAfterBadPID(t *testing.T) {
	t.Parallel()
	self := strconv.Itoa(os.Getpid())
	_, errBuf, err := run("-s", "0", "notapid", self)
	if err == nil {
		t.Fatal("expected failure because one operand is invalid")
	}
	if !strings.Contains(errBuf, "notapid") {
		t.Errorf("stderr = %q, want invalid pid reported", errBuf)
	}
}
