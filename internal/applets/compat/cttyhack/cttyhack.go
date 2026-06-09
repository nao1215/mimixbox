// Package cttyhack implements a compatibility cttyhack: BusyBox's cttyhack
// allocates a controlling terminal and then execs a program. MimixBox cannot
// perform that controlling-TTY trick portably, so this is an honest exec
// wrapper that forwards stdio and documents the limitation rather than
// pretending the TTY hack succeeded.
package cttyhack

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cttyhack applet.
type Command struct{}

// New returns a cttyhack command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cttyhack" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run PROGRAM with the current stdio (no controlling-TTY trick)" }

// Run executes the program operand, forwarding the shell's streams.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "PROGRAM [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Run PROGRAM with MimixBox's standard input, output, and error. Unlike BusyBox " +
			"cttyhack it does not allocate a new controlling terminal; it is a plain exec wrapper.",
		Examples: []command.Example{
			{Command: "cttyhack sh", Explain: "Run sh with the current stdio."},
		},
		Notes: []string{
			"The controlling-TTY allocation that BusyBox cttyhack performs is not implemented.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	operands := fs.Args()
	if len(operands) == 0 {
		return command.Failuref("missing PROGRAM operand")
	}

	path, argv := resolve(operands)
	cmd := exec.CommandContext(ctx, path, argv...) //nolint:gosec // running a user-named program is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// resolve maps the program operand to the path and argv to run: a PATH binary
// when present, otherwise this MimixBox binary re-invoked as the applet.
func resolve(operands []string) (string, []string) {
	if _, err := exec.LookPath(operands[0]); err == nil {
		return operands[0], operands[1:]
	}
	if self, err := os.Executable(); err == nil {
		return self, operands
	}
	return operands[0], operands[1:]
}
