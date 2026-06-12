// Package chvt implements the chvt applet: switch the foreground to a given
// virtual terminal.
package chvt

import (
	"context"
	"os"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the chvt applet.
type Command struct{}

// New returns a chvt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chvt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Switch to a virtual terminal" }

// VT_ACTIVATE / VT_WAITACTIVE ioctls, not exported by this x/sys version.
const (
	vtActivate   = 0x5606
	vtWaitActive = 0x5607
)

// switchFn is indirected so the VT switch can be tested without a console.
var switchFn = func(n int) error {
	f, err := os.Open("/dev/console") //nolint:gosec // the system console
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), vtActivate, uintptr(n)); errno != 0 {
		return errno
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), vtWaitActive, uintptr(n)); errno != 0 {
		return errno
	}
	return nil
}

// Run executes chvt.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "N", stdio.Err).WithHelp(command.Help{
		Description: "Switch the foreground virtual terminal to N (activating it and waiting until it " +
			"is active), via the VT_ACTIVATE/VT_WAITACTIVE ioctls on the system console. Requires a " +
			"Linux virtual console and privilege.",
		Examples: []command.Example{
			{Command: "chvt 2", Explain: "Switch to tty2."},
		},
		ExitStatus: "0  the terminal was switched.\n1  no valid N was given or the switch failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a virtual terminal number is required")
	}
	n, err := strconv.Atoi(rest[0])
	if err != nil || n < 1 {
		return command.Failuref("invalid virtual terminal number: %q", rest[0])
	}

	if err := switchFn(n); err != nil {
		return command.Failuref("cannot switch to vt %d: %v", n, err)
	}
	return nil
}
