// Package deallocvt implements the deallocvt applet: deallocate unused virtual
// terminals.
package deallocvt

import (
	"context"
	"os"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the deallocvt applet.
type Command struct{}

// New returns a deallocvt command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "deallocvt" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Deallocate a virtual terminal" }

// vtDisallocate is the VT_DISALLOCATE ioctl, not exported by this x/sys version.
const vtDisallocate = 0x5608

// deallocFn is indirected so deallocation can be tested without a console.
var deallocFn = func(n int) error {
	f, err := os.Open("/dev/console") //nolint:gosec // the system console
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), vtDisallocate, uintptr(n)); errno != 0 {
		return errno
	}
	return nil
}

// Run executes deallocvt.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[N]", stdio.Err).WithHelp(command.Help{
		Description: "Deallocate the unused virtual terminal N (via the VT_DISALLOCATE ioctl on the " +
			"system console). With no argument, or 0, deallocate all unused virtual terminals. A VT " +
			"that is in use cannot be deallocated. Requires a Linux virtual console and privilege.",
		Examples: []command.Example{
			{Command: "deallocvt 3", Explain: "Deallocate tty3 if it is unused."},
			{Command: "deallocvt", Explain: "Deallocate all unused VTs."},
		},
		ExitStatus: "0  the terminal(s) were deallocated.\n1  an invalid N, or the ioctl failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	n := 0 // 0 means "all unused"
	if rest := fs.Args(); len(rest) > 0 {
		if n, err = strconv.Atoi(rest[0]); err != nil || n < 0 {
			return command.Failuref("invalid virtual terminal number: %q", rest[0])
		}
	}

	if err := deallocFn(n); err != nil {
		return command.Failuref("cannot deallocate vt %d: %v", n, err)
	}
	return nil
}
