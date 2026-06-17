// Package timeout implements the timeout applet: run a command and terminate it
// if it is still running after a given duration.
package timeout

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/signal"
)

// Command is the timeout applet.
type Command struct{}

// New returns a timeout command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "timeout" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a command with a time limit" }

// Exit codes follow GNU timeout's convention.
const (
	exitTimedOut  = 124
	exitCannotRun = 126
	exitNotFound  = 127
)

// Run executes timeout.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION] DURATION COMMAND [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND and terminate it if it is still running after DURATION. DURATION is a " +
			"number with an optional suffix: s for seconds (the default), m for minutes, h for hours, " +
			"or d for days.",
		Examples: []command.Example{
			{Command: "timeout 10 sleep 30", Explain: "Run \"sleep 30\" but kill it after ten seconds."},
			{Command: "timeout -s KILL 5 ./worker", Explain: "Send SIGKILL instead of SIGTERM after five seconds."},
		},
		ExitStatus: "0    success.\n124  COMMAND timed out and was terminated.\n" +
			"126  COMMAND was found but could not be run.\n127  COMMAND was not found.",
	})
	// Stop parsing options at the first operand (the DURATION) so flags meant
	// for the wrapped command are passed through untouched.
	fs.SetInterspersed(false)
	signalName := fs.StringP("signal", "s", "TERM", "specify the signal to send on timeout")
	killAfter := fs.StringP("kill-after", "k", "", "also send KILL if still running this long after the initial signal")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("missing duration and/or command")
	}

	dur, err := parseDuration(rest[0])
	if err != nil {
		return command.Failuref("invalid time interval %q", rest[0])
	}
	sig, err := parseSignal(*signalName)
	if err != nil {
		return command.Failuref("%v", err)
	}

	var killGrace time.Duration
	if *killAfter != "" {
		if killGrace, err = parseDuration(*killAfter); err != nil {
			return command.Failuref("invalid time interval %q", *killAfter)
		}
	}

	return c.runWithTimeout(stdio, dur, killGrace, sig, rest[1], rest[2:])
}

// runWithTimeout starts name with args and sends sig to it if it is still alive
// after dur, mapping the result to GNU timeout's exit codes. When killGrace > 0
// and the process is still running that long after the initial signal, SIGKILL
// is sent (GNU timeout's -k/--kill-after).
func (c *Command) runWithTimeout(stdio command.IO, dur, killGrace time.Duration, sig syscall.Signal, name string, args []string) error {
	cmd := exec.Command(name, args...) //nolint:gosec // running a user-named command is the point
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	// Run the command in its own process group so that, on timeout, the whole
	// group can be signalled instead of only the direct child. Otherwise a
	// wrapper that backgrounds work (e.g. a shell script) would leak its
	// descendants after timeout returns (issue #951).
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return &command.ExitError{Code: exitNotFound, Err: fmt.Errorf("%s: command not found", name)}
		}
		return &command.ExitError{Code: exitCannotRun, Err: fmt.Errorf("%s: %v", name, err)}
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-timer.C:
		signalGroup(cmd.Process.Pid, sig)
		if killGrace > 0 {
			select {
			case <-done:
			case <-time.After(killGrace):
				signalGroup(cmd.Process.Pid, syscall.SIGKILL)
				<-done
			}
		} else {
			<-done
		}
		return &command.ExitError{Code: exitTimedOut}
	case err := <-done:
		if err == nil {
			return nil
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return &command.ExitError{Code: exitCannotRun, Err: err}
	}
}

// signalGroup sends sig to the entire process group led by pid. The child is
// started with Setpgid, so its PID is also its process-group ID; a negative
// PID addresses the whole group (see kill(2)). If the group is already gone the
// resulting ESRCH is ignored. As a safety net, the direct child is signalled
// too in case the group could not be addressed.
func signalGroup(pid int, sig syscall.Signal) {
	if err := syscall.Kill(-pid, sig); err != nil {
		_ = syscall.Kill(pid, sig)
	}
}

// parseDuration parses a GNU timeout duration: a number with an optional
// s/m/h/d suffix; a bare number means seconds.
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	unit := time.Second
	switch s[len(s)-1] {
	case 's':
		unit, s = time.Second, s[:len(s)-1]
	case 'm':
		unit, s = time.Minute, s[:len(s)-1]
	case 'h':
		unit, s = time.Hour, s[:len(s)-1]
	case 'd':
		unit, s = 24*time.Hour, s[:len(s)-1]
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 {
		return 0, fmt.Errorf("invalid")
	}
	return time.Duration(f * float64(unit)), nil
}

// parseSignal resolves a signal given by name (with or without the SIG prefix)
// or by number to the corresponding syscall.Signal, using the canonical signal
// table.
func parseSignal(name string) (syscall.Signal, error) {
	n, err := signal.NumberLax(name)
	if err != nil {
		return 0, err
	}
	return syscall.Signal(n), nil
}
