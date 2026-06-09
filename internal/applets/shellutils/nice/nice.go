// Package nice implements the nice applet: run a command with an adjusted
// scheduling priority (niceness), or print the current niceness.
package nice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the nice applet.
type Command struct{}

// New returns a nice command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nice" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a command with an adjusted niceness" }

// These are indirected so tests can observe them without touching the host.
// getpriority returns the actual niceness (-20..19): the Linux getpriority
// syscall returns 20-nice to stay non-negative, so it is converted back here.
var (
	getpriority = func() (int, error) {
		raw, err := unix.Getpriority(unix.PRIO_PROCESS, 0)
		return 20 - raw, err
	}
	setpriority = func(nice int) error { return unix.Setpriority(unix.PRIO_PROCESS, 0, nice) }
)

// Run executes nice.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n ADJUST] [COMMAND [ARG]...]", stdio.Err).WithHelp(command.Help{
		Description: "Run COMMAND with its niceness increased by ADJUST (default 10; higher means " +
			"lower priority). With no COMMAND, print the current niceness.",
		Examples: []command.Example{
			{Command: "nice", Explain: "Print the current niceness."},
			{Command: "nice -n 5 sort big.txt", Explain: "Run sort at a lower priority."},
		},
		ExitStatus: "0  success.\n1  the adjustment or command was invalid.\n127  COMMAND was not found.",
	})
	adjust := fs.IntP("adjustment", "n", 10, "add ADJUST to the niceness")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		cur, gerr := getpriority()
		if gerr != nil {
			return command.Failuref("cannot read niceness: %v", gerr)
		}
		_, _ = fmt.Fprintln(stdio.Out, cur)
		return nil
	}

	cur, gerr := getpriority()
	if gerr != nil {
		cur = 0
	}
	if serr := setpriority(cur + *adjust); serr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "nice: cannot set niceness: %v\n", serr)
	}

	cmd := exec.CommandContext(ctx, rest[0], rest[1:]...) //nolint:gosec // running a user-named command is the point
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &command.ExitError{Code: exitErr.ExitCode()}
		}
		if errors.Is(err, exec.ErrNotFound) || os.IsNotExist(err) {
			_, _ = fmt.Fprintf(stdio.Err, "nice: %s: command not found\n", rest[0])
			return &command.ExitError{Code: 127}
		}
		return command.Failuref("%v", err)
	}
	return nil
}
