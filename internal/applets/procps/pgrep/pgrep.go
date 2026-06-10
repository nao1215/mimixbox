// Package pgrep implements the pgrep and pkill applets: find (and, for pkill,
// signal) processes whose name matches a pattern.
package pgrep

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/signal"
)

// Command is the pgrep or pkill applet.
type Command struct{ kill bool }

// NewPgrep returns the pgrep applet.
func NewPgrep() *Command { return &Command{} }

// NewPkill returns the pkill applet.
func NewPkill() *Command { return &Command{kill: true} }

// Name returns the command name.
func (c *Command) Name() string {
	if c.kill {
		return "pkill"
	}
	return "pgrep"
}

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.kill {
		return "Signal processes by name"
	}
	return "Find process IDs by name"
}

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// sendSignal is indirected so signalling is testable.
var sendSignal = func(pid int, sig syscall.Signal) error { return syscall.Kill(pid, sig) }

// Run executes pgrep/pkill.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	usage := "PATTERN"
	if c.kill {
		usage = "[-SIGNAL] PATTERN"
	}
	fs := command.NewFlagSet(c.Name(), usage, stdio.Err).WithHelp(c.help())
	sigSpec := fs.StringP("signal", "s", "TERM", "signal to send (pkill)")

	// Let pkill accept the "-9"/"-KILL" shorthand by rewriting it to --signal.
	if c.kill && len(args) > 0 && strings.HasPrefix(args[0], "-") && args[0] != "--help" && args[0] != "--version" {
		if _, err := signal.Number(strings.TrimPrefix(args[0], "-")); err == nil {
			args = append([]string{"--signal", strings.TrimPrefix(args[0], "-")}, args[1:]...)
		}
	}

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: a pattern is required\n", c.Name())
		return command.SilentFailure()
	}
	re, err := regexp.Compile(rest[0])
	if err != nil {
		return command.Failuref("invalid pattern: %v", err)
	}

	pids := match(re)
	if len(pids) == 0 {
		return command.SilentFailure() // pgrep/pkill exit 1 when nothing matches
	}

	if !c.kill {
		for _, pid := range pids {
			_, _ = fmt.Fprintln(stdio.Out, pid)
		}
		return nil
	}

	sigNum, err := signal.Number(*sigSpec)
	if err != nil {
		return command.Failuref("%v", err)
	}
	failed := false
	for _, pid := range pids {
		if err := sendSignal(pid, syscall.Signal(sigNum)); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "pkill: signalling %d failed: %v\n", pid, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// match returns the sorted PIDs whose process name matches re.
func match(re *regexp.Regexp) []int {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join(procDir, e.Name(), "comm")) //nolint:gosec // /proc path
		if err != nil {
			continue
		}
		if re.MatchString(strings.TrimSpace(string(data))) {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	return pids
}

func (c *Command) help() command.Help {
	if c.kill {
		return command.Help{
			Description: "Send a signal (SIGTERM by default, or -SIGNAL / -s SIGNAL) to every process " +
				"whose name matches the regular expression PATTERN.",
			Examples: []command.Example{
				{Command: "pkill -9 firefox", Explain: "Force-kill processes named like firefox."},
			},
			ExitStatus: "0  at least one process was signalled.\n1  nothing matched.",
		}
	}
	return command.Help{
		Description: "Print the process IDs of every process whose name matches the regular expression " +
			"PATTERN, one per line in ascending order.",
		Examples: []command.Example{
			{Command: "pgrep sshd", Explain: "List the PIDs of sshd processes."},
		},
		ExitStatus: "0  at least one process matched.\n1  nothing matched.",
	}
}
