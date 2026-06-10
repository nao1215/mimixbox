// Package killall5 implements the killall5 applet: send a signal to all
// processes except kernel threads, PID 1, and the caller itself.
package killall5

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/signal"
)

// Command is the killall5 applet.
type Command struct{}

// New returns a killall5 command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "killall5" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Send a signal to all processes" }

// Injected so the destructive behavior is testable.
var (
	procDir    = "/proc"
	selfPid    = os.Getpid
	sendSignal = func(pid int, sig syscall.Signal) error { return syscall.Kill(pid, sig) }
)

// Run executes killall5.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-SIGNAL] [-o PID]...", stdio.Err).WithHelp(command.Help{
		Description: "Send a signal (SIGTERM by default, or -SIGNAL) to every process except kernel " +
			"threads, PID 1, the killall5 process itself, and any PID excluded with -o. This is the " +
			"SysV-init tool used during shutdown.",
		Examples: []command.Example{
			{Command: "killall5 -15", Explain: "Send SIGTERM to all processes."},
			{Command: "killall5 -9 -o 1234", Explain: "SIGKILL all but PID 1234."},
		},
		ExitStatus: "0  a signal was sent.\n2  no processes were found to signal.",
	})
	var omit []int
	fs.IntSliceVarP(&omit, "omit", "o", nil, "do not signal this PID")

	// Accept the "-15"/"-TERM" shorthand by rewriting it to --signal.
	sigSpec := "TERM"
	var passthrough []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") && a != "--help" && a != "--version" {
			if _, err := signal.Number(strings.TrimPrefix(a, "-")); err == nil {
				sigSpec = strings.TrimPrefix(a, "-")
				continue
			}
		}
		passthrough = append(passthrough, a)
	}

	proceed, err := fs.Parse(stdio, passthrough)
	if err != nil || !proceed {
		return err
	}

	sigNum, err := signal.Number(sigSpec)
	if err != nil {
		return command.Failuref("%v", err)
	}

	excluded := map[int]bool{1: true, c.self(): true}
	for _, p := range omit {
		excluded[p] = true
	}

	targets := c.targets(excluded)
	if len(targets) == 0 {
		return &command.ExitError{Code: 2}
	}
	for _, pid := range targets {
		_ = sendSignal(pid, syscall.Signal(sigNum)) // a process may exit before we reach it
	}
	return nil
}

func (c *Command) self() int { return selfPid() }

// targets returns the sorted PIDs to signal: every process with a command line
// (i.e. not a kernel thread) that is not excluded.
func (c *Command) targets(excluded map[int]bool) []int {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || excluded[pid] {
			continue
		}
		// Kernel threads have an empty cmdline; skip them.
		data, err := os.ReadFile(filepath.Join(procDir, e.Name(), "cmdline")) //nolint:gosec // /proc path
		if err != nil || len(data) == 0 {
			continue
		}
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	return pids
}
