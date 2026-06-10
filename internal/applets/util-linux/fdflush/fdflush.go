// Package fdflush implements the fdflush applet: flush a floppy device's buffers
// so the next access re-reads the medium.
package fdflush

import (
	"context"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fdflush applet.
type Command struct{}

// New returns a fdflush command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fdflush" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Flush a floppy device's buffers" }

// fdFlush is the FDFLUSH ioctl, not exported by this x/sys version.
const fdFlush = 0x254b

// flushFn is indirected so the privileged ioctl can be tested.
var flushFn = func(device string) error {
	f, err := os.Open(device) //nolint:gosec // user-named device
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fdFlush, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// Run executes fdflush.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Force the kernel to forget its cached view of the floppy DEVICE, so the next " +
			"access re-reads the medium (useful after swapping a disk). Requires privilege.",
		Examples: []command.Example{
			{Command: "fdflush /dev/fd0", Explain: "Flush the buffers of /dev/fd0."},
		},
		ExitStatus: "0  the buffers were flushed.\n1  no device was given or the ioctl failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a floppy device is required")
	}

	if err := flushFn(rest[0]); err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}
