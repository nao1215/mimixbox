// Package pwdx implements the pwdx applet: print the current working directory
// of one or more processes.
package pwdx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pwdx applet.
type Command struct{}

// New returns a pwdx command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pwdx" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the working directory of a process" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// Run executes pwdx.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "PID...", stdio.Err).WithHelp(command.Help{
		Description: "Print the current working directory of each process given by PID, read from " +
			"/proc/PID/cwd.",
		Examples: []command.Example{
			{Command: "pwdx 1234", Explain: "Print process 1234's working directory."},
		},
		ExitStatus: "0  every PID was resolved.\n1  a PID was invalid or could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	pids := fs.Args()
	if len(pids) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "pwdx: a PID is required")
		return command.SilentFailure()
	}

	failed := false
	for _, p := range pids {
		if _, err := strconv.Atoi(p); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "pwdx: %s: invalid process id\n", p)
			failed = true
			continue
		}
		cwd, err := os.Readlink(filepath.Join(procDir, p, "cwd"))
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "pwdx: %s: %v\n", p, err)
			failed = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s: %s\n", p, cwd)
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
