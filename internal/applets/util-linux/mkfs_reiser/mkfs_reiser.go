// Package mkfsreiser implements the mkfs.reiser applet. ReiserFS is deprecated
// and being removed from Linux, so this build does not create one; it fails
// deterministically with an explanation rather than writing an invalid image.
package mkfsreiser

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mkfs.reiser applet.
type Command struct{}

// New returns a mkfs.reiser command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkfs.reiser" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a ReiserFS filesystem (unsupported)" }

// Run executes mkfs.reiser.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Create a ReiserFS filesystem. ReiserFS is deprecated and is being removed from " +
			"the Linux kernel, so this build does not create one: rather than writing an unverifiable " +
			"image it fails with this explanation. Use mke2fs (ext2), mkfs.vfat (FAT), or mkfs.minix " +
			"instead.",
		Examples: []command.Example{
			{Command: "mkfs.reiser disk.img", Explain: "Reports that ReiserFS is unsupported."},
		},
		ExitStatus: "1  always: ReiserFS creation is not supported by this build.",
		Notes: []string{
			"ReiserFS is deprecated and slated for removal from the Linux kernel; create a supported filesystem (such as ext4) instead.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if len(fs.Args()) == 0 {
		return command.Failuref("a device or image is required")
	}

	_, _ = fmt.Fprintln(stdio.Err, "mkfs.reiser: ReiserFS is deprecated and being removed from Linux; "+
		"this build does not create ReiserFS filesystems. Use mke2fs, mkfs.vfat, or mkfs.minix instead.")
	return command.SilentFailure()
}
