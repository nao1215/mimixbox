// Package pidof implements the pidof applet: find the process IDs of running
// programs by name.
package pidof

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pidof applet.
type Command struct{}

// New returns a pidof command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pidof" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Find the process ID of a running program" }

// process is one running process as far as pidof cares.
type process struct {
	pid  int
	name string
}

// listProcesses enumerates the running processes; tests replace it.
var listProcesses = procFromProcfs

// Run executes pidof.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... PROGRAM...", stdio.Err)
	single := fs.BoolP("single-shot", "s", false, "return only one PID")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		return command.Failuref("missing program name")
	}

	procs, err := listProcesses()
	if err != nil {
		return command.Failuref("cannot list processes: %v", err)
	}

	pids := matchPIDs(procs, names, *single)
	if len(pids) == 0 {
		// pidof exits 1 with no output when nothing matches.
		return &command.ExitError{Code: command.ExitFailure}
	}

	parts := make([]string, len(pids))
	for i, p := range pids {
		parts[i] = strconv.Itoa(p)
	}
	if _, err := fmt.Fprintln(stdio.Out, strings.Join(parts, " ")); err != nil {
		return command.Failure(err)
	}
	return nil
}

// matchPIDs returns the PIDs whose process name matches any requested program,
// in descending PID order (newest first), like pidof. With single, only the
// first match is returned.
func matchPIDs(procs []process, names []string, single bool) []int {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[filepath.Base(n)] = true
	}
	var pids []int
	for _, p := range procs {
		if want[p.name] {
			pids = append(pids, p.pid)
		}
	}
	// procFromProcfs already returns descending PIDs; honour single-shot.
	if single && len(pids) > 1 {
		pids = pids[:1]
	}
	return pids
}

// procFromProcfs reads the running processes from /proc.
func procFromProcfs() ([]process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var procs []process
	// Walk in reverse so the highest PID (newest) comes first.
	for i := len(entries) - 1; i >= 0; i-- {
		pid, err := strconv.Atoi(entries[i].Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/proc", entries[i].Name(), "comm"))
		if err != nil {
			continue
		}
		procs = append(procs, process{pid: pid, name: strings.TrimSpace(string(data))})
	}
	return procs, nil
}
