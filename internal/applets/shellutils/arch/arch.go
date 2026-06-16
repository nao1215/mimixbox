// Package arch implements the arch applet: print the machine hardware name,
// equivalent to "uname -m".
package arch

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the arch applet.
type Command struct{}

// New returns an arch command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "arch" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print machine hardware name (same as uname -m)" }

// machine is the source of the hardware name; tests replace it.
var machine = realMachine

// Run executes arch.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the machine hardware name, equivalent to \"uname -m\".",
		Examples: []command.Example{
			{Command: "arch", Explain: "print the hardware name, e.g. x86_64"},
		},
		ExitStatus: "0  success.\n1  the machine hardware name could not be determined.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name, err := machine()
	if err != nil {
		return command.Failuref("cannot get machine hardware name: %v", err)
	}
	if _, err := fmt.Fprintln(stdio.Out, name); err != nil {
		return command.Failure(err)
	}
	return nil
}

// realMachine reads the machine field from the uname(2) system call.
func realMachine() (string, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return "", err
	}
	n := 0
	for n < len(u.Machine) && u.Machine[n] != 0 {
		n++
	}
	return string(u.Machine[:n]), nil
}
