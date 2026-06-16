// Package nohup implements the nohup applet: run a command so it ignores the
// hang-up signal, redirecting its output to nohup.out when stdout is a terminal.
package nohup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/term"
)

// Command is the nohup applet.
type Command struct{}

// New returns a nohup command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nohup" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a command immune to hangups, with output to a non-tty" }

// Exit codes follow GNU nohup's convention.
const (
	exitCannotRun = 126
	exitNotFound  = 127
)

// isTerminal reports whether w is a terminal; tests replace it.
var isTerminal = func(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(int(f.Fd()))
}

// Run executes nohup.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "COMMAND [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND with the hang-up signal (SIGHUP) ignored, so it keeps running after the controlling terminal is closed. When standard output is a terminal it is redirected, appending to nohup.out (or $HOME/nohup.out).",
		Examples: []command.Example{
			{Command: "nohup ./long-job.sh &", Explain: "Run long-job.sh so it survives logout, with output in nohup.out."},
			{Command: "nohup make build", Explain: "Run 'make build' immune to hangups."},
		},
		ExitStatus: "0    COMMAND ran and exited 0.\n126  COMMAND was found but could not be run.\n127  COMMAND was not found.",
	})
	// Stop parsing options at the command name so its flags pass through.
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("missing command operand")
	}

	out, cleanup, err := outputWriter(stdio)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer cleanup()

	// Ignore SIGHUP for the duration so a terminal hang-up does not kill us
	// (and, via inheritance, the child) before the command finishes.
	signal.Ignore(syscall.SIGHUP)
	defer signal.Reset(syscall.SIGHUP)

	cmd := exec.Command(rest[0], rest[1:]...) //nolint:gosec // running a user-named command is the point
	cmd.Stdin = stdio.In
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		switch {
		case errors.Is(err, exec.ErrNotFound):
			return &command.ExitError{Code: exitNotFound, Err: fmt.Errorf("%s: command not found", rest[0])}
		case errors.As(err, &ee):
			return &command.ExitError{Code: ee.ExitCode()}
		default:
			return &command.ExitError{Code: exitCannotRun, Err: fmt.Errorf("%s: %v", rest[0], err)}
		}
	}
	return nil
}

// outputWriter returns where the command's output should go. When stdout is a
// terminal nohup appends to nohup.out (or $HOME/nohup.out) like GNU nohup;
// otherwise the existing stdout is used. The returned cleanup closes any file
// that was opened.
func outputWriter(stdio command.IO) (io.Writer, func(), error) {
	if !isTerminal(stdio.Out) {
		return stdio.Out, func() {}, nil
	}
	f, err := os.OpenFile("nohup.out", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		if home := os.Getenv("HOME"); home != "" {
			f, err = os.OpenFile(home+"/nohup.out", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		}
	}
	if err != nil {
		return nil, func() {}, fmt.Errorf("cannot open nohup.out: %v", err)
	}
	_, _ = fmt.Fprintln(stdio.Err, "nohup: ignoring input and appending output to 'nohup.out'")
	return f, func() { _ = f.Close() }, nil
}
