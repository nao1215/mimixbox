// Package eject implements the eject applet: eject removable media (open or
// close a CD-ROM tray).
package eject

import (
	"context"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the eject applet.
type Command struct{}

// New returns an eject command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "eject" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Eject removable media" }

// CD-ROM ioctls, not exported by this x/sys version.
const (
	cdromEject     = 0x5309
	cdromCloseTray = 0x5319
)

// defaultDevice is the device used when none is given.
const defaultDevice = "/dev/cdrom"

// ejectFn is indirected so the privileged ioctl can be tested.
var ejectFn = func(device string, closeTray bool) error {
	f, err := os.Open(device) //nolint:gosec // user-named device
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	op := uintptr(cdromEject)
	if closeTray {
		op = cdromCloseTray
	}
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), op, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// Run executes eject.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t] [DEVICE]", stdio.Err).WithHelp(command.Help{
		Description: "Eject removable media by opening the tray of the CD-ROM DEVICE (default " +
			"/dev/cdrom). With -t, close the tray instead. Operating the drive requires privilege.",
		Examples: []command.Example{
			{Command: "eject", Explain: "Open the /dev/cdrom tray."},
			{Command: "eject -t /dev/sr0", Explain: "Close the tray of /dev/sr0."},
		},
		ExitStatus: "0  the tray was operated.\n1  the device could not be opened or the ioctl failed.",
	})
	closeTray := fs.BoolP("trayclose", "t", false, "close the tray instead of opening it")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	device := defaultDevice
	if rest := fs.Args(); len(rest) > 0 {
		device = rest[0]
	}

	if err := ejectFn(device, *closeTray); err != nil {
		return command.Failuref("%s: %v", device, err)
	}
	return nil
}
