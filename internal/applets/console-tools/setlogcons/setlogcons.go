// Package setlogcons implements the setlogcons applet: send kernel messages to a
// given virtual terminal.
package setlogcons

import (
	"context"
	"os"
	"strconv"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the setlogcons applet.
type Command struct{}

// New returns a setlogcons command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setlogcons" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Send kernel messages to a VT" }

// TIOCLINUX subcommand 11 sets the console that receives kernel messages.
const (
	tioclinux    = 0x541C
	setLogSubcmd = 11
)

// setFn is indirected so the ioctl can be tested without a console.
var setFn = func(n int) error {
	f, err := os.Open("/dev/console") //nolint:gosec // the system console
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	arg := [2]byte{setLogSubcmd, byte(n)}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), tioclinux, uintptr(unsafe.Pointer(&arg))); errno != 0 {
		return errno
	}
	return nil
}

// Run executes setlogcons.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[N]", stdio.Err).WithHelp(command.Help{
		Description: "Direct kernel log messages to virtual terminal N (via the TIOCLINUX ioctl). With " +
			"no argument, or 0, send them to the current foreground terminal. Requires privilege and a " +
			"Linux console.",
		Examples: []command.Example{
			{Command: "setlogcons 12", Explain: "Send kernel messages to tty12."},
		},
		ExitStatus: "0  the log console was set.\n1  an invalid N, or the ioctl failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	n := 0
	if rest := fs.Args(); len(rest) > 0 {
		if n, err = strconv.Atoi(rest[0]); err != nil || n < 0 {
			return command.Failuref("invalid virtual terminal number: %q", rest[0])
		}
	}

	if err := setFn(n); err != nil {
		return command.Failuref("cannot set the log console: %v", err)
	}
	return nil
}
