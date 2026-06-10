// Package mesg implements the mesg applet: display or control write access to
// the controlling terminal (whether other users may send it messages with
// write/wall).
package mesg

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the mesg applet.
type Command struct{}

// New returns a mesg command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mesg" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display or control write access to your terminal" }

// resolveTTY returns the path of the terminal on r; tests replace it.
var resolveTTY = func(r io.Reader) (string, error) {
	f, ok := r.(*os.File)
	if !ok {
		return "", os.ErrInvalid
	}
	if _, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS); err != nil {
		return "", err
	}
	return os.Readlink(fmt.Sprintf("/proc/self/fd/%d", f.Fd()))
}

// groupWrite is the terminal's group-write permission bit.
const groupWrite = 0o020

// Run executes mesg.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[y|n]", stdio.Err).WithHelp(command.Help{
		Description: "With no argument, report whether other users may write to your terminal " +
			"('is y' or 'is n'). 'mesg y' grants access and 'mesg n' revokes it by toggling the " +
			"terminal's group-write permission.",
		Examples: []command.Example{
			{Command: "mesg", Explain: "Report the current state."},
			{Command: "mesg n", Explain: "Disallow messages to your terminal."},
		},
		ExitStatus: "0  access is allowed.\n1  access is denied.\n2  an error occurred (e.g. not a terminal).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	path, err := resolveTTY(stdio.In)
	if err != nil {
		_, _ = fmt.Fprintln(stdio.Err, "mesg: cannot get terminal name")
		return &command.ExitError{Code: 2}
	}
	info, err := os.Stat(path)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mesg: %s\n", command.FileError(path, err))
		return &command.ExitError{Code: 2}
	}
	allowed := info.Mode()&groupWrite != 0

	operands := fs.Args()
	if len(operands) == 0 {
		if allowed {
			_, _ = fmt.Fprintln(stdio.Out, "is y")
			return nil
		}
		_, _ = fmt.Fprintln(stdio.Out, "is n")
		return &command.ExitError{Code: 1}
	}
	if len(operands) > 1 {
		_, _ = fmt.Fprintln(stdio.Err, "mesg: too many arguments")
		return &command.ExitError{Code: 2}
	}

	switch operands[0] {
	case "y":
		if err := os.Chmod(path, info.Mode()|groupWrite); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mesg: %v\n", err)
			return &command.ExitError{Code: 2}
		}
		return nil
	case "n":
		if err := os.Chmod(path, info.Mode()&^groupWrite); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mesg: %v\n", err)
			return &command.ExitError{Code: 2}
		}
		return &command.ExitError{Code: 1}
	default:
		_, _ = fmt.Fprintf(stdio.Err, "mesg: invalid argument: %q\n", operands[0])
		return &command.ExitError{Code: 2}
	}
}
