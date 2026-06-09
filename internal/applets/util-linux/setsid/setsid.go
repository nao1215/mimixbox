// Package setsid implements the setsid applet: run a program in a new session,
// detached from the calling process's controlling terminal.
package setsid

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the setsid applet.
type Command struct{}

// New returns a setsid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setsid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program in a new session" }

// Run executes setsid.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c] PROGRAM [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Run PROGRAM in a new session (setsid(2)), so it becomes a session and process " +
			"group leader with no controlling terminal. With -c, also set the controlling terminal " +
			"to the current one when standard input is a terminal.",
		Examples: []command.Example{
			{Command: "setsid mydaemon", Explain: "Start mydaemon detached from this terminal."},
		},
		ExitStatus: "The exit status of PROGRAM (127 if it could not be run).",
	})
	fs.SetInterspersed(false)
	ctty := fs.BoolP("ctty", "c", false, "set the controlling terminal to the current one")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "setsid: missing PROGRAM operand")
		return command.SilentFailure()
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running a user-named program is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if *ctty {
		cmd.SysProcAttr.Setctty = true
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		_, _ = fmt.Fprintf(stdio.Err, "setsid: %s: %v\n", rest[0], err)
		return &command.ExitError{Code: 127}
	}
	return nil
}
