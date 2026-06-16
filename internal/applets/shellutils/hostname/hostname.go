// Package hostname implements the hostname applet: print the system's host
// name, optionally trimmed to the short form.
package hostname

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the hostname applet.
type Command struct{}

// New returns a hostname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "hostname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the system's host name" }

// hostFn is the source of the host name; tests replace it.
var hostFn = os.Hostname

// Run executes hostname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the name of the current host. With -s, print only the short host " +
			"name by trimming everything from the first dot onward.",
		Examples: []command.Example{
			{Command: "hostname", Explain: "Print the full host name."},
			{Command: "hostname -s", Explain: "Print the short host name (up to the first dot)."},
		},
		ExitStatus: "0  the host name was printed.\n1  the host name could not be determined.",
	})
	short := fs.BoolP("short", "s", false, "display the short host name (up to the first dot)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name, err := hostFn()
	if err != nil {
		return command.Failuref("cannot determine host name: %v", err)
	}
	if *short {
		name, _, _ = strings.Cut(name, ".")
	}
	if _, err := fmt.Fprintln(stdio.Out, name); err != nil {
		return command.Failure(err)
	}
	return nil
}
