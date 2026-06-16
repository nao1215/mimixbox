// Package logname implements the logname applet: print the login name of the
// current user.
package logname

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the logname applet.
type Command struct{}

// New returns a logname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "logname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the name of the current user" }

// loginName resolves the current login name; tests replace it.
var loginName = currentLogin

// Run executes logname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the login name of the current user, taken from the LOGNAME or USER environment variables, falling back to the current account name.",
		Examples: []command.Example{
			{Command: "logname", Explain: "Print the current login name, e.g. 'alice'."},
		},
		ExitStatus: "0  the login name was printed.\n1  no login name could be determined.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name, err := loginName()
	if err != nil {
		return command.Failuref("no login name")
	}
	if _, err := fmt.Fprintln(stdio.Out, name); err != nil {
		return command.Failure(err)
	}
	return nil
}

// currentLogin returns the login name from the LOGNAME or USER environment
// variables, falling back to the current user's account name.
func currentLogin() (string, error) {
	for _, key := range []string{"LOGNAME", "USER"} {
		if v := os.Getenv(key); v != "" {
			return v, nil
		}
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}
