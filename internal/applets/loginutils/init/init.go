// Package initapplet implements the init applet (also linuxrc): run the one-shot
// startup actions of an inittab. The full PID-1 supervision loop is not provided.
package initapplet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the init applet. It is also registered under the name linuxrc.
type Command struct{ name string }

// New returns an init command.
func New() *Command { return &Command{name: "init"} }

// NewLinuxrc returns the same applet under the name linuxrc.
func NewLinuxrc() *Command { return &Command{name: "linuxrc"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run an inittab's startup actions" }

// inittabPath is the init configuration; tests point it at a fixture.
var inittabPath = "/etc/inittab"

// runFn runs one inittab process, waiting for it when wait is true.
var runFn = func(stdio command.IO, process string, wait bool) error {
	cmd := exec.Command("sh", "-c", process) //nolint:gosec // running configured init actions is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if wait {
		return cmd.Run()
	}
	return cmd.Start()
}

// oneShotActions are run during startup; respawn/askfirst entries would need the
// supervision loop and are skipped by this slice.
var oneShotActions = map[string]bool{"sysinit": true, "wait": true, "once": true}

// waitActions block until the process finishes.
var waitActions = map[string]bool{"sysinit": true, "wait": true}

// Run executes init.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t INITTAB]", stdio.Err).WithHelp(command.Help{
		Description: "Read an inittab and run its one-shot startup actions — sysinit and wait entries " +
			"(run to completion) and once entries (started in the background) — in file order. The " +
			"respawn/askfirst entries and the full PID-1 supervision loop of a real init are not run by " +
			"this build. Each inittab line is 'id:runlevels:action:process'.",
		Examples: []command.Example{
			{Command: "init -t /etc/inittab", Explain: "Run the inittab's startup actions."},
		},
		ExitStatus: "0  the startup actions ran.\n1  the inittab could not be read.",
	})
	tab := fs.StringP("inittab", "t", "", "inittab file (default /etc/inittab)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *tab != "" {
		inittabPath = *tab
	}

	data, err := os.ReadFile(inittabPath) //nolint:gosec // well-known inittab path
	if err != nil {
		return command.Failuref("cannot read %s: %v", inittabPath, err)
	}

	for _, entry := range parseInittab(string(data)) {
		if !oneShotActions[entry.action] {
			continue
		}
		if err := runFn(stdio, entry.process, waitActions[entry.action]); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "init: %s: %v\n", entry.process, err)
		}
	}
	return nil
}

type entry struct {
	action  string
	process string
}

// parseInittab parses inittab text into id:runlevels:action:process entries,
// skipping blanks and comments.
func parseInittab(text string) []entry {
	var entries []entry
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.SplitN(line, ":", 4)
		if len(fields) != 4 {
			continue
		}
		entries = append(entries, entry{action: fields[2], process: fields[3]})
	}
	return entries
}
