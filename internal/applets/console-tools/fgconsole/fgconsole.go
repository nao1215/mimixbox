// Package fgconsole implements the fgconsole applet: print the number of the
// active virtual terminal.
package fgconsole

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the fgconsole applet.
type Command struct{}

// New returns a fgconsole command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fgconsole" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the active virtual terminal" }

// vtGetState is the VT_GETSTATE ioctl, not exported by this x/sys version.
const vtGetState = 0x5603

// vtStat mirrors struct vt_stat: the active VT and the signal/state masks.
type vtStat struct {
	active, signal, state uint16
}

// getActiveFn is indirected so the active VT can be tested without a console.
var getActiveFn = func() (int, error) {
	f, err := os.Open("/dev/tty") //nolint:gosec // the controlling terminal
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	var vs vtStat
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), vtGetState, uintptr(unsafe.Pointer(&vs))); errno != 0 {
		return 0, errno
	}
	return int(vs.active), nil
}

// Run executes fgconsole.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the number of the virtual terminal that is currently in the foreground, " +
			"read with the VT_GETSTATE ioctl on the controlling terminal. Requires a Linux virtual " +
			"console.",
		Examples: []command.Example{
			{Command: "fgconsole", Explain: "Print e.g. '1' for tty1."},
		},
		ExitStatus: "0  the active VT was printed.\n1  the console state could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	active, err := getActiveFn()
	if err != nil {
		return command.Failuref("cannot read the console state: %v", err)
	}
	_, _ = fmt.Fprintln(stdio.Out, active)
	return nil
}
