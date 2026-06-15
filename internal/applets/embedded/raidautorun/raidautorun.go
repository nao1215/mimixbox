// Package raidautorun implements the raidautorun applet: tell the kernel to
// auto-detect and start RAID arrays via an md device.
package raidautorun

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the raidautorun applet.
type Command struct{}

// New returns a raidautorun command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "raidautorun" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Auto-detect and start RAID arrays" }

// autorun issues the RAID_AUTORUN ioctl on an md device. It is injected so the
// operand handling is testable without an md device or privilege.
var autorun AutoRunner = osAutoRunner{}

// AutoRunner abstracts the privileged RAID_AUTORUN ioctl.
type AutoRunner interface {
	// Run triggers RAID autodetection through the md control device.
	Run(device string) error
}

// Run executes raidautorun.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Tell the kernel md driver to scan partitions of type 0xFD and assemble any RAID arrays " +
			"it finds, using DEVICE (an md control device such as /dev/md0) as the entry point. This starts " +
			"arrays and requires privilege; without it the command fails with a documented error.",
		Examples: []command.Example{
			{Command: "raidautorun /dev/md0", Explain: "Auto-assemble RAID arrays via /dev/md0."},
		},
		ExitStatus: "0  autodetection was triggered.\n1  bad arguments or the ioctl was denied.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 1 {
		return command.Failuref("usage: raidautorun DEVICE")
	}
	if err := autorun.Run(rest[0]); err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	return nil
}
