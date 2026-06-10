// Package freeramdisk implements the freeramdisk applet: free the memory used by
// a ramdisk device.
package freeramdisk

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
	"os"
)

// Command is the freeramdisk applet.
type Command struct{}

// New returns a freeramdisk command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "freeramdisk" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Free the memory used by a ramdisk" }

// blkFlsBuf is the BLKFLSBUF ioctl, not exported by this x/sys version.
const blkFlsBuf = 0x1261

// flushFn is indirected so the privileged ioctl can be tested.
var flushFn = func(device string) error {
	f, err := os.OpenFile(device, os.O_RDWR, 0) //nolint:gosec // user-named device
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), blkFlsBuf, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// Run executes freeramdisk.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Free (flush the buffers of) the ramdisk block DEVICE, returning its memory to the " +
			"system. The device must not be in use, and the operation requires privilege.",
		Examples: []command.Example{
			{Command: "freeramdisk /dev/ram0", Explain: "Free the ramdisk /dev/ram0."},
		},
		ExitStatus: "0  the ramdisk was freed.\n1  no device was given or the ioctl failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a ramdisk device is required")
	}

	if err := flushFn(rest[0]); err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}
