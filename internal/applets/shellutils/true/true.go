// Package booltrue implements the true applet: do nothing, successfully.
package booltrue

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the true applet.
type Command struct{}

// New returns a true command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "true" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Do nothing. Return success(0)" }

// Run always succeeds. Like GNU true it ignores its operands, except that a
// leading --help or --version is honored.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "--help":
			command.NewFlagSet(c.Name(), "[IGNORED]...", stdio.Err).WriteUsage(stdio.Out)
		case "--version":
			_, _ = command.NewFlagSet(c.Name(), "", stdio.Err).Parse(stdio, []string{"--version"})
		}
	}
	return nil
}
