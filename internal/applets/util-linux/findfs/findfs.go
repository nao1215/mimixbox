// Package findfs implements the findfs applet: find a block device by its
// filesystem LABEL or UUID.
package findfs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the findfs applet.
type Command struct{}

// New returns a findfs command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "findfs" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Find a filesystem by label or UUID" }

// The udev symlink directories; tests point these at fixtures.
var (
	byLabelDir = "/dev/disk/by-label"
	byUUIDDir  = "/dev/disk/by-uuid"
)

// Run executes findfs.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "LABEL=<label>|UUID=<uuid>", stdio.Err).WithHelp(command.Help{
		Description: "Print the block device whose filesystem has the given LABEL or UUID, resolved " +
			"through /dev/disk/by-label and /dev/disk/by-uuid.",
		Examples: []command.Example{
			{Command: "findfs LABEL=rootfs", Explain: "Print the device labeled rootfs."},
			{Command: "findfs UUID=1234-5678", Explain: "Print the device with that UUID."},
		},
		ExitStatus: "0  the device was found.\n1  the tag could not be resolved.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "findfs: a LABEL=<label> or UUID=<uuid> tag is required")
		return command.SilentFailure()
	}
	spec := rest[0]

	kind, value, ok := strings.Cut(spec, "=")
	if !ok {
		return c.unresolved(stdio, spec)
	}

	var dir string
	switch strings.ToUpper(kind) {
	case "LABEL":
		dir = byLabelDir
	case "UUID":
		dir = byUUIDDir
	default:
		return c.unresolved(stdio, spec)
	}

	device, err := resolve(filepath.Join(dir, value))
	if err != nil {
		return c.unresolved(stdio, spec)
	}
	_, _ = fmt.Fprintln(stdio.Out, device)
	return nil
}

func (c *Command) unresolved(stdio command.IO, spec string) error {
	_, _ = fmt.Fprintf(stdio.Err, "findfs: unable to resolve '%s'\n", spec)
	return command.SilentFailure()
}

// resolve reads the udev symlink and returns the absolute device path.
func resolve(link string) (string, error) {
	target, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(target) {
		return filepath.Clean(target), nil
	}
	return filepath.Clean(filepath.Join(filepath.Dir(link), target)), nil
}
