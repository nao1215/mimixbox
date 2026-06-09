// Package taskset implements the taskset applet: retrieve or set the CPU
// affinity of a process, or launch a command bound to a set of CPUs.
package taskset

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the taskset applet.
type Command struct{}

// New returns a taskset command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "taskset" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Set or get a process's CPU affinity" }

// Indirected so affinity changes can be tested without touching real processes.
var (
	getAffinity = func(pid int, set *unix.CPUSet) error { return unix.SchedGetaffinity(pid, set) }
	setAffinity = func(pid int, set *unix.CPUSet) error { return unix.SchedSetaffinity(pid, set) }
)

// Run executes taskset.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c] [-p] MASK|CPU-LIST [COMMAND|PID]...", stdio.Err).WithHelp(command.Help{
		Description: "Without -p, run COMMAND bound to the given CPUs and pass its flags through. " +
			"With -p, print a PID's current affinity, or set it when a mask is given first. The CPUs " +
			"are a hexadecimal MASK, or a CPU-LIST like '0,2-4' when -c is used.",
		Examples: []command.Example{
			{Command: "taskset -c 0,1 -- make", Explain: "Run make on CPUs 0 and 1."},
			{Command: "taskset -p 1234", Explain: "Print process 1234's affinity mask."},
		},
		ExitStatus: "0  success.\n1  an invalid mask/PID, or the command could not be run.",
	})
	fs.SetInterspersed(false)
	pidMode := fs.BoolP("pid", "p", false, "operate on an existing PID")
	cpuList := fs.BoolP("cpu-list", "c", false, "interpret the operand as a CPU list, not a hex mask")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	operands := fs.Args()

	if *pidMode {
		return c.pidMode(stdio, operands, *cpuList)
	}
	return c.runMode(ctx, stdio, operands, *cpuList)
}

// pidMode handles "-p PID" (print) and "-p MASK PID" (set).
func (c *Command) pidMode(stdio command.IO, operands []string, cpuList bool) error {
	switch len(operands) {
	case 1:
		pid, err := strconv.Atoi(operands[0])
		if err != nil {
			return command.Failuref("invalid PID: %q", operands[0])
		}
		var set unix.CPUSet
		if err := getAffinity(pid, &set); err != nil {
			return command.Failuref("failed to get affinity for %d: %v", pid, err)
		}
		_, _ = fmt.Fprintf(stdio.Out, "pid %d's current affinity mask: %s\n", pid, maskHex(&set))
		return nil
	case 2:
		set, err := parseSet(operands[0], cpuList)
		if err != nil {
			return command.Failuref("%v", err)
		}
		pid, err := strconv.Atoi(operands[1])
		if err != nil {
			return command.Failuref("invalid PID: %q", operands[1])
		}
		if err := setAffinity(pid, set); err != nil {
			return command.Failuref("failed to set affinity for %d: %v", pid, err)
		}
		_, _ = fmt.Fprintf(stdio.Out, "pid %d's new affinity mask: %s\n", pid, maskHex(set))
		return nil
	default:
		_, _ = fmt.Fprintln(stdio.Err, "taskset: -p needs a PID (optionally preceded by a mask)")
		return command.SilentFailure()
	}
}

// runMode handles "MASK COMMAND" / "-c CPU-LIST COMMAND".
func (c *Command) runMode(ctx context.Context, stdio command.IO, operands []string, cpuList bool) error {
	if len(operands) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "taskset: a mask and a command are required")
		return command.SilentFailure()
	}
	set, err := parseSet(operands[0], cpuList)
	if err != nil {
		return command.Failuref("%v", err)
	}
	rest := operands[1:]
	if rest[0] == "--" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "taskset: a command is required")
		return command.SilentFailure()
	}
	if err := setAffinity(0, set); err != nil {
		return command.Failuref("failed to set affinity: %v", err)
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

// parseSet builds a CPUSet from a hex mask or a CPU list ("0,2-4").
func parseSet(s string, cpuList bool) (*unix.CPUSet, error) {
	var set unix.CPUSet
	if cpuList {
		for _, part := range strings.Split(s, ",") {
			if part == "" {
				continue
			}
			if lo, hi, ok := strings.Cut(part, "-"); ok {
				a, err1 := strconv.Atoi(lo)
				b, err2 := strconv.Atoi(hi)
				if err1 != nil || err2 != nil || a > b {
					return nil, fmt.Errorf("invalid CPU list: %q", s)
				}
				for i := a; i <= b; i++ {
					set.Set(i)
				}
			} else {
				n, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("invalid CPU list: %q", s)
				}
				set.Set(n)
			}
		}
		return &set, nil
	}

	v := new(big.Int)
	hex := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if _, ok := v.SetString(hex, 16); !ok {
		return nil, fmt.Errorf("invalid mask: %q", s)
	}
	for i := 0; i < v.BitLen(); i++ {
		if v.Bit(i) == 1 {
			set.Set(i)
		}
	}
	return &set, nil
}

// maskHex renders a CPUSet as the lowercase hex bitmask taskset prints.
func maskHex(set *unix.CPUSet) string {
	v := new(big.Int)
	for i := 0; i < 1024; i++ {
		if set.IsSet(i) {
			v.SetBit(v, i, 1)
		}
	}
	if v.Sign() == 0 {
		return "0"
	}
	return v.Text(16)
}
