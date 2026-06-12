// Package setconsole implements the setconsole applet: redirect system console
// output to a given device.
package setconsole

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
	"os"
)

// Command is the setconsole applet.
type Command struct{}

// New returns a setconsole command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setconsole" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Redirect console output to a device" }

// tioccons redirects console output to the terminal on which it is issued.
const tioccons = 0x541D

// defaultDevice resets the console to the real console.
const defaultDevice = "/dev/console"

// redirectFn is indirected so the ioctl can be tested without a console.
var redirectFn = func(device string) error {
	f, err := os.Open(device) //nolint:gosec // user-named console device
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), tioccons, 0); errno != 0 {
		return errno
	}
	return nil
}

// Run executes setconsole.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-r] [DEVICE]", stdio.Err).WithHelp(command.Help{
		Description: "Redirect kernel console output to DEVICE (the current terminal by default) using " +
			"the TIOCCONS ioctl. With -r, reset the redirection back to /dev/console. Requires " +
			"privilege.",
		Examples: []command.Example{
			{Command: "setconsole /dev/ttyS0", Explain: "Send console output to the serial port."},
			{Command: "setconsole -r", Explain: "Reset console output to /dev/console."},
		},
		ExitStatus: "0  the redirection was set.\n1  the device was inaccessible or the ioctl failed.",
	})
	reset := fs.BoolP("reset", "r", false, "reset console output to /dev/console")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	device := defaultDevice
	if *reset {
		device = defaultDevice
	} else if rest := fs.Args(); len(rest) > 0 {
		device = rest[0]
	}

	if err := redirectFn(device); err != nil {
		return command.Failuref("cannot redirect the console to %s: %v", device, err)
	}
	return nil
}
