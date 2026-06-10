// Package fstrim implements the fstrim applet: discard unused blocks on a
// mounted filesystem.
package fstrim

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fstrim applet.
type Command struct{}

// New returns a fstrim command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fstrim" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Discard unused blocks on a filesystem" }

// fiTrim is the FITRIM ioctl, not exported by this x/sys version.
const fiTrim = 0xC0185879

// fstrimRange mirrors struct fstrim_range.
type fstrimRange struct {
	start  uint64
	length uint64
	minlen uint64
}

// trimFn is indirected so the privileged ioctl can be tested. It returns the
// number of bytes the kernel reported as trimmed.
var trimFn = func(path string) (uint64, error) {
	f, err := os.Open(path) //nolint:gosec // user-named mount point
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()

	r := fstrimRange{start: 0, length: ^uint64(0), minlen: 0}
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), fiTrim, uintptr(unsafe.Pointer(&r)))
	if errno != 0 {
		return 0, errno
	}
	return r.length, nil
}

// Run executes fstrim.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-v] MOUNTPOINT", stdio.Err).WithHelp(command.Help{
		Description: "Discard blocks that are not in use by the filesystem mounted at MOUNTPOINT, " +
			"telling the underlying device (SSD/thin volume) it may reclaim them. With -v, report how " +
			"many bytes were trimmed. The operation requires privilege.",
		Examples: []command.Example{
			{Command: "fstrim /", Explain: "Trim the root filesystem."},
			{Command: "fstrim -v /home", Explain: "Trim /home and report the amount."},
		},
		ExitStatus: "0  the filesystem was trimmed.\n1  no mount point was given or the ioctl failed.",
	})
	verbose := fs.BoolP("verbose", "v", false, "report the number of bytes trimmed")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a mount point is required")
	}
	mountpoint := rest[0]

	trimmed, err := trimFn(mountpoint)
	if err != nil {
		return command.Failuref("%s: %v", mountpoint, err)
	}
	if *verbose {
		_, _ = fmt.Fprintf(stdio.Out, "%s: %d bytes were trimmed\n", mountpoint, trimmed)
	}
	return nil
}
