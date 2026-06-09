// Package bracket implements the "[" and "[[" test aliases: thin compatibility
// fronts over the test applet that require a closing "]" / "]]" and otherwise
// share test's exit-status semantics (0 true, 1 false, 2 malformed).
package bracket

import (
	"context"
	"errors"

	testcmd "github.com/nao1215/mimixbox/internal/applets/shellutils/test"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the "[" or "[[" alias; close holds the required closing token.
type Command struct {
	name  string
	close string
}

// NewBracket returns the "[" alias.
func NewBracket() *Command { return &Command{name: "[", close: "]"} }

// NewDoubleBracket returns the "[[" alias.
func NewDoubleBracket() *Command { return &Command{name: "[[", close: "]]"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Evaluate a conditional expression (test alias requiring " + c.close + ")"
}

// Run validates the closing bracket and delegates to test.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	// --help / --version remain discoverable even though POSIX "[" has no options.
	if len(args) > 0 && (args[0] == "--help" || args[0] == "--version") {
		fs := command.NewFlagSet(c.Name(), "EXPRESSION "+c.close, stdio.Err).WithHelp(command.Help{
			Description: "Evaluate EXPRESSION as the test command does, but require a closing " +
				c.close + ". The result is the exit status only.",
			Examples: []command.Example{
				{Command: c.Name() + " -f /etc/hosts " + c.close, Explain: "True when the file exists and is regular."},
				{Command: c.Name() + ` "$x" = y ` + c.close, Explain: "True when $x equals y."},
			},
			ExitStatus: "0  the expression is true.\n1  the expression is false.\n2  the expression is malformed.",
		})
		_, _ = fs.Parse(stdio, args)
		return nil
	}

	if len(args) == 0 || args[len(args)-1] != c.close {
		return &command.ExitError{Code: 2, Err: errors.New("missing " + c.close)}
	}
	return testcmd.New().Run(ctx, stdio, args[:len(args)-1])
}
