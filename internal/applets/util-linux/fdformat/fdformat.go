// Package fdformat implements the fdformat applet: low-level format a floppy
// device, track by track.
package fdformat

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fdformat applet.
type Command struct{}

// New returns a fdformat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fdformat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Low-level format a floppy device" }

// Floppy ioctls, not exported by this x/sys version.
const (
	fdGetPrm = 0x0204 // FDGETPRM: read the drive geometry
	fdFmtBeg = 0x0247 // FDFMTBEG: begin a format session
	fdFmtTrk = 0x0248 // FDFMTTRK: format one track
	fdFmtEnd = 0x0249 // FDFMTEND: end a format session
)

// floppyStruct mirrors the kernel's struct floppy_struct (FDGETPRM). The C
// fields are 32-bit, so they are uint32 here, not Go's 64-bit uint.
type floppyStruct struct {
	Size, Sect, Head, Track, Stretch uint32
	Gap, Rate, Spec1, FmtGap         uint8
	Name                             uintptr
}

// formatDescr mirrors struct format_descr (FDFMTTRK): the device field is
// unused, and head/track select the track to format.
type formatDescr struct {
	Device, Head, Track uint32
}

// formatFn is indirected so the privileged format sequence can be tested.
var formatFn = func(device string) (tracks int, err error) {
	f, err := os.Open(device) //nolint:gosec // user-named device
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	var g floppyStruct
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fdGetPrm, uintptr(unsafe.Pointer(&g))); errno != 0 {
		return 0, errno
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fdFmtBeg, 0); errno != 0 {
		return 0, errno
	}
	heads := g.Head
	if heads == 0 {
		heads = 1
	}
	cylinders := g.Size / (g.Sect * heads)
	for cyl := uint32(0); cyl < cylinders; cyl++ {
		for head := uint32(0); head < heads; head++ {
			fd := formatDescr{Head: head, Track: cyl}
			if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fdFmtTrk, uintptr(unsafe.Pointer(&fd))); errno != 0 {
				return 0, errno
			}
		}
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fdFmtEnd, 0); errno != 0 {
		return 0, errno
	}
	return int(cylinders * heads), nil
}

// Run executes fdformat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Low-level format the floppy DEVICE: read its geometry, then format every track. " +
			"This erases the medium and writes fresh sector headers; it does not create a filesystem " +
			"(use mkfs.* afterwards). It requires a real floppy drive and privilege.",
		Examples: []command.Example{
			{Command: "fdformat /dev/fd0", Explain: "Format the floppy in /dev/fd0."},
		},
		ExitStatus: "0  the floppy was formatted.\n1  no device was given or the format failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a floppy device is required")
	}

	_, _ = fmt.Fprintf(stdio.Out, "Formatting %s ...\n", rest[0])
	tracks, err := formatFn(rest[0])
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "done (%d tracks)\n", tracks)
	return nil
}
