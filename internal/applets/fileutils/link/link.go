// Package link implements the link applet: create a hard link to a file by
// calling the link system call, like the GNU link utility.
package link

import (
	"context"
	"errors"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the link applet.
type Command struct{}

// New returns a link command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "link" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a hard link to a file" }

// Run executes link.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE1 FILE2", stdio.Err).WithHelp(command.Help{
		Description: "Create a hard link named FILE2 to the existing file FILE1 by calling the " +
			"link system call directly, with none of the options of the ln command.",
		Examples: []command.Example{
			{Command: "link file.txt hardlink.txt", Explain: "Create a hard link hardlink.txt to file.txt."},
		},
		ExitStatus: "0  the link was created.\n1  the link could not be created or the operands were wrong.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) != 2 {
		return command.Failuref("exactly two arguments are required")
	}
	if err := os.Link(names[0], names[1]); err != nil {
		return command.Failuref("cannot create link %q to %q: %v", names[1], names[0], unwrap(err))
	}
	return nil
}

// unwrap reduces an *os.LinkError to its underlying reason so the message does
// not repeat the file names that the command already prints.
func unwrap(err error) error {
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err
	}
	return err
}
