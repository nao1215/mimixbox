// Package shellfront implements the BusyBox shell front-ends sh, ash, hush and
// bash as compatibility launchers over MimixBox's own shell, mbsh. They accept
// the common invocation forms (interactive, -c COMMAND, -s, SCRIPT [ARG]...) and
// only show an interactive prompt when standard input is a terminal.
package shellfront

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mbsh"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is one shell front-end (its name distinguishes sh/ash/hush/bash).
type Command struct{ name string }

// NewSh returns the sh front-end.
func NewSh() *Command { return &Command{name: "sh"} }

// NewAsh returns the ash front-end.
func NewAsh() *Command { return &Command{name: "ash"} }

// NewHush returns the hush front-end.
func NewHush() *Command { return &Command{name: "hush"} }

// NewBash returns the bash front-end.
func NewBash() *Command { return &Command{name: "bash"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Command interpreter (MimixBox mbsh compatibility front-end)"
}

// Run launches mbsh according to the requested invocation form.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c COMMAND | -s | SCRIPT [ARG]...]", stdio.Err).WithHelp(command.Help{
		Description: "A compatibility front-end over MimixBox's shell (mbsh). It runs a one-off " +
			"COMMAND with -c, reads commands from standard input with -s or no operand, or runs " +
			"a SCRIPT file. An interactive prompt is shown only when standard input is a terminal.",
		Examples: []command.Example{
			{Command: c.Name() + " -c 'echo hello'", Explain: "Run a single command and exit."},
			{Command: c.Name() + " script.sh", Explain: "Run the commands in script.sh."},
			{Command: "echo 'echo hi' | " + c.Name(), Explain: "Read commands from standard input."},
		},
		Notes: []string{
			"This is mbsh under another name, not a full BusyBox/POSIX shell; it supports mbsh's syntax.",
		},
	})
	commandStr := fs.StringP("command", "c", "", "run COMMAND and exit")
	fromStdin := fs.BoolP("stdin", "s", false, "read commands from standard input")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	operands := fs.Args()

	switch {
	case fs.Changed("command"):
		in := strings.NewReader(*commandStr + "\n")
		mbsh.Interpret(ctx, command.IO{In: in, Out: stdio.Out, Err: stdio.Err}, false)
	case *fromStdin || len(operands) == 0:
		mbsh.Interpret(ctx, stdio, isTerminal(stdio.In))
	default:
		data, rerr := os.ReadFile(operands[0]) //nolint:gosec // user-named script
		if rerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %s\n", c.Name(), command.FileError(operands[0], rerr))
			return command.SilentFailure()
		}
		mbsh.Interpret(ctx, command.IO{In: bytes.NewReader(data), Out: stdio.Out, Err: stdio.Err}, false)
	}
	return nil
}

// isTerminal reports whether r is a character device (a real terminal).
func isTerminal(r interface{}) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
