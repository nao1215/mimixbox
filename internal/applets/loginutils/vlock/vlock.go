// Package vlock implements the vlock applet: lock the terminal until the user's
// password is entered.
package vlock

import (
	"bufio"
	"context"
	"fmt"
	"os/user"

	"github.com/nao1215/mimixbox/internal/auth"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the vlock applet.
type Command struct{}

// New returns a vlock command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "vlock" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Lock the terminal until the password is entered" }

// Injected so the current user and the auth backend are testable.
var (
	currentUserFn = func() (string, error) {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		return u.Username, nil
	}
	authFn = auth.Authenticate
)

// Run executes vlock.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [-c]", stdio.Err).WithHelp(command.Help{
		Description: "Lock the terminal: print a notice and unlock only when the current user's password " +
			"is entered on standard input. -a (lock all consoles) and -c (clear the screen) are " +
			"accepted for compatibility. The password is verified against the system authentication " +
			"backend.",
		Examples: []command.Example{
			{Command: "vlock", Explain: "Lock the terminal until the password is entered."},
		},
		ExitStatus: "0  the correct password unlocked the terminal.\n1  the password was wrong.",
	})
	_ = fs.BoolP("all", "a", false, "lock all consoles (accepted for compatibility)")
	_ = fs.BoolP("current", "c", false, "lock the current console (the default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	username, err := currentUserFn()
	if err != nil {
		return command.Failuref("cannot determine the current user: %v", err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "This TTY is now locked.\nPlease enter the password for %s to unlock: ", username)
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		return command.Failuref("no password entered")
	}

	ok, err := authFn(username, sc.Text())
	if err != nil {
		return command.Failuref("%v", err)
	}
	if !ok {
		return command.Failuref("incorrect password")
	}
	_, _ = fmt.Fprintln(stdio.Out, "\nTerminal unlocked.")
	return nil
}
