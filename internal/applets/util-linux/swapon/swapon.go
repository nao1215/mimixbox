// Package swapon implements the swapon applet: enable a swap area, or list the
// active ones.
package swapon

import (
	"context"
	"fmt"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the swapon applet.
type Command struct{}

// New returns a swapon command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "swapon" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Enable a swap area or list active swaps" }

// Injected so the privileged call and the table are testable.
var (
	swapsPath = "/proc/swaps"
	swaponFn  = func(path string) error {
		p, err := unix.BytePtrFromString(path)
		if err != nil {
			return err
		}
		_, _, errno := unix.Syscall(unix.SYS_SWAPON, uintptr(unsafe.Pointer(p)), 0, 0)
		if errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes swapon.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-s] [FILE...]", stdio.Err).WithHelp(command.Help{
		Description: "Enable swapping on each FILE or device. With -s, or with no operand, print the " +
			"table of active swap areas from /proc/swaps instead. Enabling swap requires privilege.",
		Examples: []command.Example{
			{Command: "swapon -s", Explain: "List the active swap areas."},
			{Command: "swapon /swapfile", Explain: "Enable swapping on /swapfile."},
		},
		ExitStatus: "0  success.\n1  a swap area could not be enabled, or the table could not be read.",
	})
	summary := fs.BoolP("summary", "s", false, "print the active swap areas and exit")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if *summary || len(rest) == 0 {
		data, err := os.ReadFile(swapsPath)
		if err != nil {
			return command.Failuref("cannot read %s: %v", swapsPath, err)
		}
		_, _ = stdio.Out.Write(data)
		return nil
	}

	failed := false
	for _, path := range rest {
		if err := swaponFn(path); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "swapon: %s: %v\n", path, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
