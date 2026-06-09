// Package busybox implements the busybox multi-call front-end: "busybox --list"
// prints the applets, "busybox APPLET [ARG]..." dispatches to one, and unknown
// applets are reported by the underlying dispatch with a non-zero exit.
package busybox

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/version"
)

// Command is the busybox applet.
type Command struct{}

// New returns a busybox command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "busybox" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "BusyBox-style multi-call front-end for MimixBox applets" }

// Run dispatches to an applet by re-invoking this binary, so an applet's
// stdout/stderr/exit-code behavior is unchanged.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.flagSet(stdio).WriteUsage(stdio.Out)
		return nil
	}
	if args[0] == "--version" || args[0] == "-v" {
		version.Print(stdio.Out, c.Name())
		return nil
	}
	target := args
	if args[0] == "--list" {
		target = []string{"--list"}
	}
	return dispatch(ctx, stdio, target)
}

func (c *Command) flagSet(stdio command.IO) *command.FlagSet {
	return command.NewFlagSet(c.Name(), "[--list] | APPLET [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "The MimixBox multi-call front-end. With --list it prints every applet; " +
			"otherwise it runs APPLET with the remaining arguments, exactly as invoking the " +
			"applet directly would.",
		Examples: []command.Example{
			{Command: "busybox --list", Explain: "List all applets."},
			{Command: "busybox cat file.txt", Explain: "Run the cat applet."},
		},
	})
}

// dispatch re-invokes this binary with args and forwards the streams and exit
// status.
func dispatch(ctx context.Context, stdio command.IO, args []string) error {
	self, err := os.Executable()
	if err != nil {
		return command.Failuref("cannot locate the MimixBox binary: %v", err)
	}
	cmd := exec.CommandContext(ctx, self, args...) //nolint:gosec // dispatching to our own applets
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
