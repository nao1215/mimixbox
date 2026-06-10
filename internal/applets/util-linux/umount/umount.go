// Package umount implements the umount applet: unmount a filesystem.
package umount

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the umount applet.
type Command struct{}

// New returns a umount command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "umount" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Unmount a filesystem" }

// Injected so the privileged call and the table are testable.
var (
	mountsPath = "/proc/mounts"
	unmountFn  = func(target string) error { return unix.Unmount(target, 0) }
)

// Run executes umount.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "TARGET", stdio.Err).WithHelp(command.Help{
		Description: "Unmount the filesystem named by TARGET, given either as the mount point or as the " +
			"mounted device. The target must be currently mounted. Unmounting requires privilege.",
		Examples: []command.Example{
			{Command: "umount /mnt/usb", Explain: "Unmount the filesystem at /mnt/usb."},
			{Command: "umount /dev/sdb1", Explain: "Unmount by device."},
		},
		ExitStatus: "0  the filesystem was unmounted.\n1  the target was not mounted or could not be unmounted.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "umount: a target is required")
		return command.SilentFailure()
	}
	target := rest[0]

	mountpoint, ok := resolveMount(target)
	if !ok {
		return command.Failuref("%s: not mounted", target)
	}
	if err := unmountFn(mountpoint); err != nil {
		return command.Failuref("%s: %v", target, err)
	}
	return nil
}

// resolveMount returns the mount point matching target (by mount point or by
// device) and whether it was found.
func resolveMount(target string) (string, bool) {
	f, err := os.Open(mountsPath)
	if err != nil {
		return "", false
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		device, mountpoint := fields[0], fields[1]
		if mountpoint == target || device == target {
			return mountpoint, true
		}
	}
	return "", false
}
