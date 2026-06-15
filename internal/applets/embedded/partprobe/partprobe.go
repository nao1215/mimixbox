// Package partprobe implements the partprobe applet: ask the kernel to re-read
// the partition table of a block device.
package partprobe

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the partprobe applet.
type Command struct{}

// New returns a partprobe command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "partprobe" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Re-read the partition table of a device" }

// reread asks the kernel to re-read a device's partition table. It is injected
// so the operand handling can be tested without a real block device.
var reread ReReader = osReReader{}

// ReReader abstracts the privileged BLKRRPART ioctl.
type ReReader interface {
	// ReRead triggers a partition-table re-read on the block device.
	ReRead(device string) error
}

// Run executes partprobe.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE...", stdio.Err).WithHelp(command.Help{
		Description: "Ask the kernel to re-read the partition table of each block DEVICE so that newly added " +
			"or removed partitions are reflected without a reboot. This changes kernel state and requires " +
			"privilege; without it partprobe fails with a documented error. It does not modify the " +
			"on-disk partition table.",
		Examples: []command.Example{
			{Command: "partprobe /dev/sda", Explain: "Re-scan the partitions of /dev/sda."},
		},
		ExitStatus: "0  every device was re-read.\n1  a device could not be re-read (often a lack of privilege).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	devices := fs.Args()
	if len(devices) == 0 {
		return command.Failuref("at least one device operand is required")
	}

	failed := false
	for _, dev := range devices {
		if err := reread.ReRead(dev); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "partprobe: %s: %v\n", dev, err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
