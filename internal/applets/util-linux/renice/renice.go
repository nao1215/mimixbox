// Package renice implements the renice applet: change the scheduling priority
// (niceness) of one or more running processes.
package renice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the renice applet.
type Command struct{}

// New returns a renice command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "renice" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Alter the priority of running processes" }

// Indirected so the priority changes can be tested without touching real
// processes. getNice returns the actual niceness (-20..19).
var (
	getNice = func(pid int) (int, error) {
		raw, err := unix.Getpriority(unix.PRIO_PROCESS, pid)
		return 20 - raw, err
	}
	setNice = func(pid, nice int) error {
		return unix.Setpriority(unix.PRIO_PROCESS, pid, nice)
	}
)

// Run executes renice.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n] PRIORITY [-p] PID...", stdio.Err).WithHelp(command.Help{
		Description: "Change the niceness (scheduling priority) of the given PIDs to PRIORITY (-20 " +
			"highest .. 19 lowest). Only the owner or the superuser may lower a process's niceness.",
		Examples: []command.Example{
			{Command: "renice 10 -p 1234", Explain: "Make process 1234 lower priority."},
			{Command: "renice -n 5 1234 5678", Explain: "Renice two processes."},
		},
		ExitStatus: "0  every process was reniced.\n1  a priority or PID was invalid, or a change was denied.",
	})
	nicePtr := fs.IntP("priority", "n", 0, "the new niceness")
	pidFlag := fs.BoolP("pid", "p", false, "interpret the operands as process IDs (the default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	_ = pidFlag

	operands := fs.Args()
	priority := *nicePtr
	if !fs.Changed("priority") {
		// The first operand is the priority when -n was not given.
		if len(operands) == 0 {
			_, _ = fmt.Fprintln(stdio.Err, "renice: missing priority")
			return command.SilentFailure()
		}
		priority, err = strconv.Atoi(operands[0])
		if err != nil {
			return command.Failuref("invalid priority: %q", operands[0])
		}
		operands = operands[1:]
	}

	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "renice: no PID given")
		return command.SilentFailure()
	}

	var failed bool
	for _, p := range operands {
		pid, err := strconv.Atoi(p)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "renice: invalid PID: %q\n", p)
			failed = true
			continue
		}
		old, gerr := getNice(pid)
		if gerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "renice: failed to get priority for %d: %v\n", pid, gerr)
			failed = true
			continue
		}
		if serr := setNice(pid, priority); serr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "renice: failed to set priority for %d: %v\n", pid, serr)
			failed = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%d (process ID) old priority %d, new priority %d\n", pid, old, priority)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
