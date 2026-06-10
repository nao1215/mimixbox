// Package fsfreeze implements the fsfreeze applet: suspend or resume access to a
// mounted filesystem.
package fsfreeze

import (
	"context"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fsfreeze applet.
type Command struct{}

// New returns a fsfreeze command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fsfreeze" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Suspend or resume a filesystem" }

// FIFREEZE/FITHAW are not exported by this x/sys version.
const (
	fiFreeze = 0xC0045877
	fiThaw   = 0xC0045878
)

// freezeFn is indirected so the privileged ioctl can be tested.
var freezeFn = func(path string, freeze bool) error {
	f, err := os.Open(path) //nolint:gosec // user-named mount point
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	op := uintptr(fiThaw)
	if freeze {
		op = fiFreeze
	}
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), op, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

// Run executes fsfreeze.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{-f|-u} MOUNTPOINT", stdio.Err).WithHelp(command.Help{
		Description: "Freeze (-f) or unfreeze (-u) the filesystem mounted at MOUNTPOINT. Freezing " +
			"suspends new writes and flushes pending ones, leaving the filesystem in a consistent " +
			"state (e.g. for a snapshot) until it is unfrozen. Exactly one of -f or -u is required, " +
			"and the operation needs privilege.",
		Examples: []command.Example{
			{Command: "fsfreeze -f /mnt/data", Explain: "Freeze the filesystem at /mnt/data."},
			{Command: "fsfreeze -u /mnt/data", Explain: "Unfreeze it again."},
		},
		ExitStatus: "0  the filesystem was frozen or unfrozen.\n1  invalid options or the ioctl failed.",
	})
	freeze := fs.BoolP("freeze", "f", false, "freeze the filesystem")
	unfreeze := fs.BoolP("unfreeze", "u", false, "unfreeze the filesystem")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *freeze == *unfreeze { // neither or both
		return command.Failuref("exactly one of -f (freeze) or -u (unfreeze) is required")
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a mount point is required")
	}

	if err := freezeFn(rest[0], *freeze); err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}
