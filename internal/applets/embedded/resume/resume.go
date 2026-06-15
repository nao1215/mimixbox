// Package resume implements the resume applet: resume the system from a
// hibernation image stored on a swap device.
package resume

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the resume applet.
type Command struct{}

// New returns a resume command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "resume" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Resume from a hibernation image" }

// resumePath is the sysfs node that triggers the resume; overridable in tests.
var resumePath = "/sys/power/resume"

// resolver maps a swap device to its "major:minor" string. It is injected so
// argument handling can be tested without a real block device.
var resolver Resolver = osResolver{}

// Resolver abstracts looking up a block device's major:minor numbers.
type Resolver interface {
	// DevNumber returns the "major:minor" of the block device at path.
	DevNumber(device string) (string, error)
}

// Run executes resume.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Resume the system from a hibernation (suspend-to-disk) image stored on the swap block " +
			"DEVICE. The device's major:minor is written to /sys/power/resume, which makes the kernel try " +
			"to restore the saved image and, on success, never returns. WARNING: this aborts the current " +
			"boot to restore a previous memory snapshot; unsaved state in the current session is lost. " +
			"It requires privilege; without it the command fails with a documented error.",
		Examples: []command.Example{
			{Command: "resume /dev/sda2", Explain: "Resume from a hibernation image on /dev/sda2 (destructive to current boot)."},
		},
		ExitStatus: "0  the resume request was written (the kernel may not return on success).\n1  bad arguments or the request was denied.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 1 {
		return command.Failuref("usage: resume DEVICE")
	}
	devnum, err := resolver.DevNumber(rest[0])
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	if err := os.WriteFile(resumePath, []byte(devnum), 0o644); err != nil {
		if os.IsPermission(err) {
			return command.Failuref("%s: permission denied (resume needs privilege)", resumePath)
		}
		return command.Failuref("%s", command.FileError(resumePath, err))
	}
	_, _ = fmt.Fprintf(stdio.Out, "resume: requested resume from %s (%s)\n", rest[0], devnum)
	return nil
}
