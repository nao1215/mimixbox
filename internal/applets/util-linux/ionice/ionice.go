// Package ionice implements the ionice applet: get or set the I/O scheduling
// class and priority of a process, or run a command with a given I/O priority.
package ionice

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the ionice applet.
type Command struct{}

// New returns an ionice command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ionice" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Get or set process I/O scheduling class and priority" }

// I/O priority encoding (see Documentation/block/ioprio).
const (
	ioprioWhoProcess = 1
	ioprioClassShift = 13
	classNone        = 0
	classRealtime    = 1
	classBestEffort  = 2
	classIdle        = 3
)

// Indirected so the priority changes can be tested without real I/O scheduling.
var (
	ioprioGet = func(pid int) (int, error) {
		r, _, errno := unix.Syscall(unix.SYS_IOPRIO_GET, ioprioWhoProcess, uintptr(pid), 0)
		if errno != 0 {
			return 0, errno
		}
		return int(r), nil
	}
	ioprioSet = func(pid, ioprio int) error {
		if _, _, errno := unix.Syscall(unix.SYS_IOPRIO_SET, ioprioWhoProcess, uintptr(pid), uintptr(ioprio)); errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes ionice.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c CLASS] [-n DATA] [-p PID | COMMAND]", stdio.Err).WithHelp(command.Help{
		Description: "Get or set a process's I/O scheduling class (1 realtime, 2 best-effort, 3 idle) " +
			"and priority data (0-7, lower is higher priority). With -p PID and no -c, print the " +
			"current class; with -c, set the PID or run COMMAND with that class.",
		Examples: []command.Example{
			{Command: "ionice -p 1234", Explain: "Show process 1234's I/O class."},
			{Command: "ionice -c 3 -- backup.sh", Explain: "Run backup.sh at idle I/O priority."},
		},
		ExitStatus: "0  success.\n1  an invalid class/PID, or the command could not be run.",
	})
	fs.SetInterspersed(false)
	class := fs.IntP("class", "c", classBestEffort, "scheduling class (1 rt, 2 be, 3 idle)")
	data := fs.IntP("classdata", "n", 4, "priority data within the class (0-7)")
	pid := fs.IntP("pid", "p", 0, "act on an existing PID")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// Reject out-of-range class/priority rather than silently coercing them.
	if *class < classNone || *class > classIdle {
		return command.Failuref("invalid class: %d (expected 1 realtime, 2 best-effort, 3 idle)", *class)
	}
	if *data < 0 || *data > 7 {
		return command.Failuref("invalid priority data: %d (expected 0-7)", *data)
	}

	rest := fs.Args()
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}
	if fs.Changed("pid") && len(rest) > 0 {
		return command.Failuref("cannot run a command together with -p")
	}

	// Print mode: -p without -c.
	if fs.Changed("pid") && !fs.Changed("class") {
		v, gerr := ioprioGet(*pid)
		if gerr != nil {
			return command.Failuref("failed to get priority for %d: %v", *pid, gerr)
		}
		_, _ = fmt.Fprintln(stdio.Out, describe(v))
		return nil
	}

	ioprio := (*class << ioprioClassShift) | *data
	if *class == classIdle || *class == classNone {
		ioprio = *class << ioprioClassShift
	}

	if fs.Changed("pid") {
		if err := ioprioSet(*pid, ioprio); err != nil {
			return command.Failuref("failed to set priority for %d: %v", *pid, err)
		}
		return nil
	}

	if len(rest) == 0 {
		// No command and no PID: report this process's own class.
		v, gerr := ioprioGet(0)
		if gerr != nil {
			return command.Failuref("failed to get priority: %v", gerr)
		}
		_, _ = fmt.Fprintln(stdio.Out, describe(v))
		return nil
	}

	if err := ioprioSet(0, ioprio); err != nil {
		return command.Failuref("failed to set priority: %v", err)
	}
	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running the user's command is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}

// describe renders an ioprio value the way ionice prints it.
func describe(ioprio int) string {
	class := ioprio >> ioprioClassShift
	data := ioprio & 0x7
	switch class {
	case classRealtime:
		return fmt.Sprintf("realtime: prio %d", data)
	case classBestEffort:
		return fmt.Sprintf("best-effort: prio %d", data)
	case classIdle:
		return "idle"
	default:
		return fmt.Sprintf("none: prio %d", data)
	}
}
