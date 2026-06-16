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
	command.HandleHelpVersionWith(stdio, c.Name(), "[IGNORED]...", command.Help{
		Description: "Do nothing and exit successfully. All operands are ignored; true exists so that " +
			"shell scripts have a command that always succeeds.",
		Examples: []command.Example{
			{Command: "true && echo ok", Explain: "Print ok, because true always succeeds."},
			{Command: "while true; do :; done", Explain: "Loop forever (true never fails)."},
		},
		ExitStatus: "0  always.",
	}, args)
	return nil
}
