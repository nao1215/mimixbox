// Package boolfalse implements the false applet: do nothing, unsuccessfully.
package boolfalse

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the false applet.
type Command struct{}

// New returns a false command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "false" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Do nothing. Return failure(1)" }

// Run always fails with exit status 1. Like GNU false it ignores its operands,
// except that a leading --help or --version is honored and exits successfully.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if command.HandleHelpVersion(stdio, c.Name(), "[IGNORED]...", args) {
		return nil
	}
	return &command.ExitError{Code: command.ExitFailure}
}
