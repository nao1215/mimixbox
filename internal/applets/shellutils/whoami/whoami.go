// Package whoami implements the whoami applet: print the user name associated
// with the current effective user ID.
package whoami

import (
	"context"
	"fmt"
	"os/user"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the whoami applet.
type Command struct{}

// New returns a whoami command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "whoami" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print login user name" }

// Run executes whoami. GNU whoami takes no operands, so any extra operand is an
// error. On success it prints the effective user name followed by a newline.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if operands := fs.Args(); len(operands) > 0 {
		_, _ = fmt.Fprintf(stdio.Err, "whoami: extra operand '%s'\n", operands[0])
		return command.SilentFailure()
	}

	u, err := user.Current()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "whoami: %v\n", err)
		return command.SilentFailure()
	}
	_, _ = fmt.Fprintln(stdio.Out, u.Username)
	return nil
}
