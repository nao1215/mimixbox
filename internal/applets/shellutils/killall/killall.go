// Package killall implements the killall applet: send a signal to every process
// running a named program.
package killall

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/signal"
)

// Command is the killall applet.
type Command struct{}

// New returns a killall command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "killall" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Kill processes by name" }

// process is one running process as far as killall cares.
type process struct {
	pid  int
	name string
}

// listProcesses enumerates running processes; tests replace it.
var listProcesses = procFromProcfs

// killProcess sends sig to pid; tests replace it.
var killProcess = func(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

// Run executes killall.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... NAME...", stdio.Err)
	signalName := fs.StringP("signal", "s", "TERM", "send this signal instead of SIGTERM")
	quiet := fs.BoolP("quiet", "q", false, "do not complain if no process was killed")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		return command.Failuref("missing process name")
	}
	sig, err := parseSignal(*signalName)
	if err != nil {
		return command.Failuref("%v", err)
	}

	procs, err := listProcesses()
	if err != nil {
		return command.Failuref("cannot list processes: %v", err)
	}

	return c.killMatching(stdio, procs, names, sig, *quiet)
}

// killMatching sends sig to every process whose name matches a requested name,
// reporting unmatched names unless quiet.
func (c *Command) killMatching(stdio command.IO, procs []process, names []string, sig syscall.Signal, quiet bool) error {
	hit := make(map[string]bool)
	for _, p := range procs {
		for _, n := range names {
			if p.name == filepath.Base(n) {
				if err := killProcess(p.pid, sig); err != nil {
					_, _ = fmt.Fprintf(stdio.Err, "killall: %d: %v\n", p.pid, err)
				} else {
					hit[filepath.Base(n)] = true
				}
			}
		}
	}

	failed := false
	for _, n := range names {
		if !hit[filepath.Base(n)] {
			failed = true
			if !quiet {
				_, _ = fmt.Fprintf(stdio.Err, "%s: no process found\n", n)
			}
		}
	}
	if failed {
		return &command.ExitError{Code: command.ExitFailure}
	}
	return nil
}

// parseSignal resolves a signal given by name (with or without SIG) or number
// using the canonical signal table.
func parseSignal(name string) (syscall.Signal, error) {
	n, err := signal.NumberLax(name)
	if err != nil {
		return 0, err
	}
	return syscall.Signal(n), nil
}

// procFromProcfs reads the running processes from /proc.
func procFromProcfs() ([]process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var procs []process
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/proc", e.Name(), "comm"))
		if err != nil {
			continue
		}
		procs = append(procs, process{pid: pid, name: strings.TrimSpace(string(data))})
	}
	return procs, nil
}
