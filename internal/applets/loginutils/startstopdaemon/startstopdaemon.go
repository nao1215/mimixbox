// Package startstopdaemon implements the start-stop-daemon applet: start or
// stop a background program, tracked by a pidfile.
package startstopdaemon

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the start-stop-daemon applet.
type Command struct{}

// New returns a start-stop-daemon command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "start-stop-daemon" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Start or stop a background program" }

// Injected so the privileged process actions are testable.
var (
	isRunning = func(pid int) bool { return unix.Kill(pid, 0) == nil }
	startProc = func(path string, args []string) (int, error) {
		attr := &os.ProcAttr{Files: []*os.File{nil, nil, nil}}
		p, err := os.StartProcess(path, append([]string{path}, args...), attr)
		if err != nil {
			return 0, err
		}
		pid := p.Pid
		_ = p.Release()
		return pid, nil
	}
	signalProc = func(pid int, sig syscall.Signal) error { return unix.Kill(pid, sig) }
)

// Run executes start-stop-daemon.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{-S|-K} [-p PIDFILE] [-x EXEC] [-s SIGNAL] [-- ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Start (-S) or stop (-K) a background program tracked by a pidfile. For -S, if " +
			"-p PIDFILE names a running process nothing is done; otherwise -x EXEC is started and its " +
			"PID written to the pidfile. For -K, the process in -p PIDFILE is sent -s SIGNAL (TERM by " +
			"default) and the pidfile is removed. Arguments after -- are passed to the started program.",
		Examples: []command.Example{
			{Command: "start-stop-daemon -S -p /run/foo.pid -x /usr/bin/foo", Explain: "Start foo if not running."},
			{Command: "start-stop-daemon -K -p /run/foo.pid", Explain: "Stop foo and remove its pidfile."},
		},
		ExitStatus: "0  the action succeeded (or -S found it already running).\n1  bad options or the action failed.",
	})
	start := fs.BoolP("start", "S", false, "start the program")
	stop := fs.BoolP("stop", "K", false, "stop the program")
	pidfile := fs.StringP("pidfile", "p", "", "pidfile tracking the process")
	execPath := fs.StringP("exec", "x", "", "program to start")
	signalName := fs.StringP("signal", "s", "TERM", "signal to send when stopping")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *start == *stop {
		return command.Failuref("exactly one of -S (start) or -K (stop) is required")
	}
	if *start {
		return doStart(stdio, *pidfile, *execPath, fs.Args())
	}
	return doStop(stdio, *pidfile, *signalName)
}

func doStart(stdio command.IO, pidfile, execPath string, args []string) error {
	if execPath == "" {
		return command.Failuref("-x EXEC is required to start a program")
	}
	if pidfile != "" {
		if pid, ok := readPidfile(pidfile); ok && isRunning(pid) {
			_, _ = fmt.Fprintf(stdio.Err, "start-stop-daemon: %s is already running (pid %d)\n", execPath, pid)
			return nil
		}
	}

	pid, err := startProc(execPath, args)
	if err != nil {
		return command.Failuref("cannot start %s: %v", execPath, err)
	}
	if pidfile != "" {
		if err := os.WriteFile(pidfile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
			return command.Failuref("cannot write %s: %v", pidfile, err)
		}
	}
	return nil
}

func doStop(stdio command.IO, pidfile, signalName string) error {
	if pidfile == "" {
		return command.Failuref("-p PIDFILE is required to stop a program")
	}
	pid, ok := readPidfile(pidfile)
	if !ok || !isRunning(pid) {
		return command.Failuref("no running process found for %s", pidfile)
	}
	sig, err := parseSignal(signalName)
	if err != nil {
		return command.Failuref("%v", err)
	}
	if err := signalProc(pid, sig); err != nil {
		return command.Failuref("cannot signal pid %d: %v", pid, err)
	}
	_ = os.Remove(pidfile)
	_, _ = fmt.Fprintf(stdio.Err, "start-stop-daemon: stopped pid %d\n", pid)
	return nil
}

func readPidfile(path string) (int, bool) {
	data, err := os.ReadFile(path) //nolint:gosec // user-named pidfile
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	return pid, true
}

// signals maps the supported signal names to their numbers.
var signals = map[string]syscall.Signal{
	"TERM": syscall.SIGTERM, "KILL": syscall.SIGKILL, "HUP": syscall.SIGHUP,
	"INT": syscall.SIGINT, "QUIT": syscall.SIGQUIT, "USR1": syscall.SIGUSR1, "USR2": syscall.SIGUSR2,
}

// parseSignal accepts a signal name (with or without SIG) or a number.
func parseSignal(name string) (syscall.Signal, error) {
	name = strings.ToUpper(strings.TrimPrefix(strings.ToUpper(name), "SIG"))
	if sig, ok := signals[name]; ok {
		return sig, nil
	}
	if n, err := strconv.Atoi(name); err == nil && n > 0 {
		return syscall.Signal(n), nil
	}
	return 0, fmt.Errorf("unknown signal: %q", name)
}
