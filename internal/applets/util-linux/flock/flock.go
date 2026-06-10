// Package flock implements the flock applet: acquire an advisory lock on a file
// and run a command while holding it.
package flock

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the flock applet.
type Command struct{}

// New returns a flock command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "flock" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a command under an advisory file lock" }

// flockFn is indirected so the locking can be observed in tests.
var flockFn = func(fd int, how int) error { return unix.Flock(fd, how) }

// Run executes flock.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-sxn] [-w SECONDS] FILE COMMAND [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Take an advisory lock on FILE, then run COMMAND while holding it; the lock is " +
			"released when COMMAND exits. The lock is exclusive by default (-x), or shared with -s. " +
			"-n fails immediately if the lock is held; -w waits up to SECONDS for it. -c runs the " +
			"argument with 'sh -c'.",
		Examples: []command.Example{
			{Command: "flock /tmp/my.lock echo hi", Explain: "Run echo while holding the lock."},
			{Command: "flock -n /tmp/my.lock -c 'job'", Explain: "Skip the job if the lock is held."},
		},
		ExitStatus: "the command's exit status, or 1 if the lock could not be taken.",
	})
	shared := fs.BoolP("shared", "s", false, "take a shared lock")
	_ = fs.BoolP("exclusive", "x", false, "take an exclusive lock (the default)")
	nonblock := fs.BoolP("nonblock", "n", false, "fail rather than wait if the lock is held")
	wait := fs.IntP("wait", "w", 0, "wait at most this many seconds for the lock")
	cmdStr := fs.StringP("command", "c", "", "run this string with sh -c")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = stdio.Err.Write([]byte("flock: a FILE is required\n"))
		return command.SilentFailure()
	}
	lockPath := rest[0]
	name, cmdArgs, ok := buildCommand(*cmdStr, rest[1:])
	if !ok {
		_, _ = stdio.Err.Write([]byte("flock: a command is required\n"))
		return command.SilentFailure()
	}

	f, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE, 0o600) //nolint:gosec // user-named lock file
	if err != nil {
		return command.Failuref("cannot open %s: %v", lockPath, err)
	}
	defer func() { _ = f.Close() }()

	how := unix.LOCK_EX
	if *shared {
		how = unix.LOCK_SH
	}
	if err := acquire(int(f.Fd()), how, *nonblock, *wait); err != nil {
		return command.Failuref("cannot lock %s: %v", lockPath, err)
	}

	cmd := exec.Command(name, cmdArgs...) //nolint:gosec // user-specified command, as flock is designed to run
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// buildCommand returns the program and arguments to run, from either -c or the
// remaining operands.
func buildCommand(cmdStr string, operands []string) (name string, args []string, ok bool) {
	if cmdStr != "" {
		return "sh", []string{"-c", cmdStr}, true
	}
	if len(operands) == 0 {
		return "", nil, false
	}
	return operands[0], operands[1:], true
}

// acquire takes the lock, honoring the non-blocking and timeout options.
func acquire(fd, how int, nonblock bool, wait int) error {
	if nonblock {
		return flockFn(fd, how|unix.LOCK_NB)
	}
	if wait <= 0 {
		return flockFn(fd, how)
	}
	deadline := time.Now().Add(time.Duration(wait) * time.Second)
	for {
		err := flockFn(fd, how|unix.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) || time.Now().After(deadline) {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
}
