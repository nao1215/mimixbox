// Package pidof implements the pidof applet: find the process IDs of running
// programs by name.
package pidof

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/proctable"
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
	fs := command.NewFlagSet(c.Name(), "[OPTION]... PROGRAM...", stdio.Err).WithHelp(command.Help{
		Description: "Find the process IDs of the named running programs and print them, newest first, " +
			"on a single line separated by spaces.",
		Examples: []command.Example{
			{Command: "pidof sshd", Explain: "Print the PIDs of all running sshd processes."},
			{Command: "pidof -s nginx", Explain: "Print only one PID for nginx."},
			{Command: "pidof bash zsh", Explain: "Print the PIDs of bash and zsh processes."},
		},
		ExitStatus: "0  at least one matching process was found.\n1  no matching process was found, or an error occurred.",
	})
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
// first match is returned. The actual matching is the shared proctable backend.
func matchPIDs(procs []process, names []string, single bool) []int {
	shared := make([]proctable.Process, len(procs))
	for i, p := range procs {
		shared[i] = proctable.Process{PID: p.pid, Name: p.name}
	}
	return proctable.MatchNames(shared, names, single)
}

// procFromProcfs reads the running processes from /proc via the shared backend.
func procFromProcfs() ([]process, error) {
	shared, err := proctable.List(proctable.DefaultProcDir)
	if err != nil {
		return nil, err
	}
	procs := make([]process, len(shared))
	for i, p := range shared {
		procs[i] = process{pid: p.PID, name: p.Name}
	}
	return procs, nil
}
