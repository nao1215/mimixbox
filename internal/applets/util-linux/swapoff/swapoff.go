// Package swapoff implements the swapoff applet: disable a swap area.
package swapoff

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the swapoff applet.
type Command struct{}

// New returns a swapoff command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "swapoff" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Disable a swap area" }

// Injected so the privileged call and the table are testable.
var (
	swapsPath = "/proc/swaps"
	swapoffFn = func(path string) error {
		p, err := unix.BytePtrFromString(path)
		if err != nil {
			return err
		}
		_, _, errno := unix.Syscall(unix.SYS_SWAPOFF, uintptr(unsafe.Pointer(p)), 0, 0)
		if errno != 0 {
			return errno
		}
		return nil
	}
)

// Run executes swapoff.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [FILE...]", stdio.Err).WithHelp(command.Help{
		Description: "Disable swapping on each FILE or device. With -a, disable every swap area listed " +
			"in /proc/swaps. Disabling swap requires privilege.",
		Examples: []command.Example{
			{Command: "swapoff /swapfile", Explain: "Disable swapping on /swapfile."},
			{Command: "swapoff -a", Explain: "Disable all swap areas."},
		},
		ExitStatus: "0  success.\n1  a swap area could not be disabled.",
	})
	all := fs.BoolP("all", "a", false, "disable all swap areas in /proc/swaps")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	targets := fs.Args()
	if *all {
		targets = activeSwaps()
	}
	if len(targets) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "swapoff: a swap area or -a is required")
		return command.SilentFailure()
	}

	failed := false
	for _, path := range targets {
		if err := swapoffFn(path); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "swapoff: %s: %v\n", path, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// activeSwaps returns the device/file column of /proc/swaps (skipping the header).
func activeSwaps() []string {
	f, err := os.Open(swapsPath)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var out []string
	sc := bufio.NewScanner(f)
	first := true
	for sc.Scan() {
		if first { // skip the "Filename Type ..." header
			first = false
			continue
		}
		fields := strings.Fields(sc.Text())
		if len(fields) > 0 {
			out = append(out, fields[0])
		}
	}
	return out
}
