// Package chrt implements the chrt applet: get or set a process's real-time
// scheduling policy and priority, or run a command under one.
package chrt

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the chrt applet.
type Command struct{}

// New returns a chrt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chrt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Get or set a process's real-time scheduling attributes" }

// Linux scheduling policies (not all exported by x/sys/unix here).
const (
	schedOther = 0
	schedFIFO  = 1
	schedRR    = 2
	schedBatch = 3
	schedIdle  = 5
)

// Indirected so the scheduling changes can be tested without affecting the test
// process.
var (
	getScheduler = func(pid int) (int, error) {
		r, _, errno := unix.Syscall(unix.SYS_SCHED_GETSCHEDULER, uintptr(pid), 0, 0)
		if errno != 0 {
			return 0, errno
		}
		return int(r), nil
	}
	getParam = func(pid int) (int, error) {
		var p int32
		if _, _, errno := unix.Syscall(unix.SYS_SCHED_GETPARAM, uintptr(pid), uintptr(unsafe.Pointer(&p)), 0); errno != 0 {
			return 0, errno
		}
		return int(p), nil
	}
	setScheduler = func(pid, policy, priority int) error {
		p := int32(priority)
		if _, _, errno := unix.Syscall(unix.SYS_SCHED_SETSCHEDULER, uintptr(pid), uintptr(policy), uintptr(unsafe.Pointer(&p))); errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes chrt.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f|-r|-o|-b|-i] [-p] PRIORITY {COMMAND|PID}", stdio.Err).WithHelp(command.Help{
		Description: "Print or change a process's scheduling policy and priority. With -p and a single " +
			"PID, print them; with a policy flag, set the PID (with -p) or run COMMAND. Policies: " +
			"-f FIFO, -r round-robin (default), -o other, -b batch, -i idle.",
		Examples: []command.Example{
			{Command: "chrt -p 1234", Explain: "Show process 1234's scheduling policy."},
			{Command: "chrt -f 50 -- server", Explain: "Run server as SCHED_FIFO priority 50."},
		},
		ExitStatus: "0  success.\n1  an invalid argument, or the command could not be run.",
	})
	fs.SetInterspersed(false)
	fifo := fs.BoolP("fifo", "f", false, "set SCHED_FIFO")
	rr := fs.BoolP("rr", "r", false, "set SCHED_RR (the default)")
	other := fs.BoolP("other", "o", false, "set SCHED_OTHER")
	batch := fs.BoolP("batch", "b", false, "set SCHED_BATCH")
	idle := fs.BoolP("idle", "i", false, "set SCHED_IDLE")
	pidFlag := fs.BoolP("pid", "p", false, "operate on an existing PID")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	policy := schedRR
	policySet := true
	switch {
	case *fifo:
		policy = schedFIFO
	case *rr:
		policy = schedRR
	case *other:
		policy = schedOther
	case *batch:
		policy = schedBatch
	case *idle:
		policy = schedIdle
	default:
		policySet = false
	}

	operands := fs.Args()
	if len(operands) > 0 && operands[0] == "--" {
		operands = operands[1:]
	}

	if *pidFlag && !policySet {
		// Print mode: chrt -p PID.
		if len(operands) != 1 {
			_, _ = fmt.Fprintln(stdio.Err, "chrt: -p without a policy needs exactly one PID")
			return command.SilentFailure()
		}
		pid, perr := strconv.Atoi(operands[0])
		if perr != nil {
			return command.Failuref("invalid PID: %q", operands[0])
		}
		return c.printPolicy(stdio, pid)
	}

	// Set mode needs a priority and a target (PID with -p, else COMMAND).
	if len(operands) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "chrt: a priority and a target are required")
		return command.SilentFailure()
	}
	priority, perr := strconv.Atoi(operands[0])
	if perr != nil {
		return command.Failuref("invalid priority: %q", operands[0])
	}

	if *pidFlag {
		pid, perr := strconv.Atoi(operands[1])
		if perr != nil {
			return command.Failuref("invalid PID: %q", operands[1])
		}
		if err := setScheduler(pid, policy, priority); err != nil {
			return command.Failuref("failed to set policy for %d: %v", pid, err)
		}
		return nil
	}

	rest := operands[1:]
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "chrt: a command is required")
		return command.SilentFailure()
	}
	if err := setScheduler(0, policy, priority); err != nil {
		return command.Failuref("failed to set policy: %v", err)
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

func (c *Command) printPolicy(stdio command.IO, pid int) error {
	policy, err := getScheduler(pid)
	if err != nil {
		return command.Failuref("failed to get policy for %d: %v", pid, err)
	}
	prio, err := getParam(pid)
	if err != nil {
		return command.Failuref("failed to get priority for %d: %v", pid, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "pid %d's current scheduling policy: %s\n", pid, policyName(policy))
	_, _ = fmt.Fprintf(stdio.Out, "pid %d's current scheduling priority: %d\n", pid, prio)
	return nil
}

func policyName(p int) string {
	switch p {
	case schedOther:
		return "SCHED_OTHER"
	case schedFIFO:
		return "SCHED_FIFO"
	case schedRR:
		return "SCHED_RR"
	case schedBatch:
		return "SCHED_BATCH"
	case schedIdle:
		return "SCHED_IDLE"
	default:
		return fmt.Sprintf("UNKNOWN (%d)", p)
	}
}
