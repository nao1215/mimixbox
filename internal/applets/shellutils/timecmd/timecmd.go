// Package timecmd implements the time applet: run a command and report how long
// it took (real, user, and system time).
package timecmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the time applet.
type Command struct{}

// New returns a time command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "time" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a command and report how long it took" }

// now is indirected so tests can supply a deterministic clock.
var now = time.Now

// Run executes time.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-p] COMMAND [ARG]...", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND with its arguments and, after it finishes, print the elapsed real, " +
			"user, and system time to standard error. With -p the times use the POSIX format.",
		Examples: []command.Example{
			{Command: "time sleep 1", Explain: "Time a command."},
			{Command: "time -p ls", Explain: "Use the portable (POSIX) output format."},
		},
		ExitStatus: "The exit status of COMMAND (127 if it could not be run).",
	})
	// Stop parsing at the command so its own flags are not consumed.
	fs.SetInterspersed(false)
	posix := fs.BoolP("portable", "p", false, "use the POSIX output format")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "time: missing command")
		return command.SilentFailure()
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running a user-named command is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err

	start := now()
	runErr := cmd.Run()
	real := now().Sub(start)

	var user, sys time.Duration
	if cmd.ProcessState != nil {
		user = cmd.ProcessState.UserTime()
		sys = cmd.ProcessState.SystemTime()
	}
	report(stdio, *posix, real, user, sys)

	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		_, _ = fmt.Fprintf(stdio.Err, "time: %s: %v\n", rest[0], runErr)
		return &command.ExitError{Code: 127}
	}
	return nil
}

// report writes the three timing lines to standard error.
func report(stdio command.IO, posix bool, real, user, sys time.Duration) {
	if posix {
		_, _ = fmt.Fprintf(stdio.Err, "real %.2f\nuser %.2f\nsys %.2f\n", real.Seconds(), user.Seconds(), sys.Seconds())
		return
	}
	_, _ = fmt.Fprintf(stdio.Err, "real\t%s\nuser\t%s\nsys\t%s\n", clock(real), clock(user), clock(sys))
}

// clock formats a duration as "<minutes>m<seconds>.<millis>s".
func clock(d time.Duration) string {
	minutes := int(d / time.Minute)
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm%.3fs", minutes, seconds)
}
