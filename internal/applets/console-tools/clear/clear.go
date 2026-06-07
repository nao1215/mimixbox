// Package clear implements the clear applet: clear the terminal screen.
package clear

import (
	"context"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// clearSequence is the escape sequence written to clear the terminal screen.
// It matches the original implementation's bytes: move the cursor to the home
// position (\033[H) and erase the display (\033[J).
const clearSequence = "\033[H\033[J"

// Command is the clear applet.
type Command struct{}

// New returns a clear command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "clear" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Clear terminal" }

// Run executes clear: it writes the terminal-clear escape sequence to stdout.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_, err = io.WriteString(stdio.Out, clearSequence)
	return err
}
