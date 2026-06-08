// Package unlink implements the unlink applet: remove a single file by calling
// the unlink system call, like the GNU unlink utility.
package unlink

import (
	"context"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the unlink applet.
type Command struct{}

// New returns an unlink command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unlink" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove a single file by calling the unlink function" }

// Run executes unlink.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) != 1 {
		return command.Failuref("exactly one argument is required")
	}
	if err := syscall.Unlink(files[0]); err != nil {
		return command.Failuref("cannot unlink %q: %v", files[0], err)
	}
	return nil
}
