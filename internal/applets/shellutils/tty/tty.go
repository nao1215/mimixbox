// Package tty implements the tty applet: print the file name of the terminal
// connected to standard input.
package tty

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/term"
)

// Command is the tty applet.
type Command struct{}

// New returns a tty command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tty" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the file name of the terminal connected to stdin" }

// fder is implemented by *os.File; it exposes the underlying descriptor so tty
// can inspect whatever standard input is wired to.
type fder interface{ Fd() uintptr }

// isTerminal reports whether fd refers to a terminal; tests replace it.
var isTerminal = term.IsTerminal

// Run executes tty.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the file name of the terminal connected to standard input. " +
			"With -s, print nothing and report the result through the exit status only.",
		Examples: []command.Example{
			{Command: "tty", Explain: "Print the terminal device, e.g. /dev/pts/0."},
			{Command: "tty -s", Explain: "Silently test whether standard input is a terminal."},
		},
		ExitStatus: "0  standard input is a terminal.\n1  standard input is not a terminal.\n2  a usage error occurred.",
	})
	silent := fs.BoolP("silent", "s", false, "print nothing, only return an exit status")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name, isTTY := ttyName(stdio.In)
	if !isTTY {
		if !*silent {
			_, _ = fmt.Fprintln(stdio.Out, "not a tty")
		}
		return &command.ExitError{Code: command.ExitFailure}
	}
	if !*silent {
		_, _ = fmt.Fprintln(stdio.Out, name)
	}
	return nil
}

// ttyName returns the terminal device path backing in, and whether in is a
// terminal at all. A non-file reader (as used by tests) is never a terminal.
func ttyName(in any) (string, bool) {
	f, ok := in.(fder)
	if !ok {
		return "", false
	}
	fd := int(f.Fd())
	if !isTerminal(fd) {
		return "", false
	}
	// /proc/self/fd/N is a symlink to the real device (e.g. /dev/pts/0).
	if link, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", fd)); err == nil {
		return link, true
	}
	return fmt.Sprintf("/dev/fd/%d", fd), true
}
