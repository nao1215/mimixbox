// Package blkdiscard implements the blkdiscard applet: discard sectors on a
// block device.
package blkdiscard

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the blkdiscard applet.
type Command struct{}

// New returns a blkdiscard command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "blkdiscard" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Discard sectors on a block device" }

// blkDiscard is the BLKDISCARD ioctl, not exported by this x/sys version.
const blkDiscard = 0x1277

// Injected so the privileged ioctls can be tested.
var (
	sizeFn = func(device string) (uint64, error) {
		f, err := os.Open(device) //nolint:gosec // user-named device
		if err != nil {
			return 0, err
		}
		defer func() { _ = f.Close() }()
		var sz uint64
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uintptr(unix.BLKGETSIZE64), uintptr(unsafe.Pointer(&sz)))
		if errno != 0 {
			return 0, errno
		}
		return sz, nil
	}
	discardFn = func(device string, offset, length uint64) error {
		f, err := os.OpenFile(device, os.O_WRONLY, 0) //nolint:gosec // user-named device
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		rng := [2]uint64{offset, length}
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), blkDiscard, uintptr(unsafe.Pointer(&rng)))
		if errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes blkdiscard.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-o OFFSET] [-l LENGTH] [-v] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Discard (tell the device to forget) a range of sectors on a block DEVICE. By " +
			"default the whole device is discarded; -o sets the byte offset to start at and -l the " +
			"number of bytes. -v reports the discarded range. This destroys data and requires privilege.",
		Examples: []command.Example{
			{Command: "blkdiscard -v /dev/sdb", Explain: "Discard the entire device."},
			{Command: "blkdiscard -o 0 -l 1048576 /dev/sdb", Explain: "Discard the first MiB."},
		},
		ExitStatus: "0  the range was discarded.\n1  no device was given or the ioctl failed.",
	})
	offset := fs.Uint64P("offset", "o", 0, "byte offset to start at")
	length := fs.Uint64P("length", "l", 0, "number of bytes to discard (0 = to end)")
	verbose := fs.BoolP("verbose", "v", false, "report the discarded range")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device is required")
	}
	device := rest[0]

	count := *length
	if count == 0 {
		size, err := sizeFn(device)
		if err != nil {
			return command.Failuref("%s: %v", device, err)
		}
		if *offset > size {
			return command.Failuref("%s: offset is past the end of the device", device)
		}
		count = size - *offset
	}

	if err := discardFn(device, *offset, count); err != nil {
		return command.Failuref("%s: %v", device, err)
	}
	if *verbose {
		_, _ = fmt.Fprintf(stdio.Out, "%s: discarded %d bytes from offset %d\n", device, count, *offset)
	}
	return nil
}
