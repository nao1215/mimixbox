// Package losetup implements the losetup applet: list active loop devices.
// Associating and detaching loop devices is privileged and not done here.
package losetup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the losetup applet.
type Command struct{}

// New returns a losetup command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "losetup" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List active loop devices" }

// sysBlockDir is the sysfs block class; tests point it at a fixture.
var sysBlockDir = "/sys/block"

// Run executes losetup.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a]", stdio.Err).WithHelp(command.Help{
		Description: "List the active loop devices and their backing files, read from /sys/block. With " +
			"-a, or with no operand, every active loop device is shown. Associating a file with a loop " +
			"device or detaching one is privileged and is not done by this build.",
		Examples: []command.Example{
			{Command: "losetup -a", Explain: "List the active loop devices."},
		},
		ExitStatus: "0  the loop devices were listed.\n1  a setup/detach was requested.",
	})
	_ = fs.BoolP("all", "a", false, "list all active loop devices (the default)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if len(fs.Args()) > 0 {
		_, _ = fmt.Fprintln(stdio.Err, "losetup: associating or detaching loop devices is not supported "+
			"by this build; run with -a to list active loop devices")
		return command.SilentFailure()
	}

	for _, d := range loopDevices() {
		_, _ = fmt.Fprintf(stdio.Out, "/dev/%s: (%s)\n", d.name, d.backing)
	}
	return nil
}

type loopDevice struct {
	name, backing string
}

// loopDevices returns the active loop devices (those with a backing file),
// sorted by name.
func loopDevices() []loopDevice {
	entries, err := os.ReadDir(sysBlockDir)
	if err != nil {
		return nil
	}
	var devices []loopDevice
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "loop") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sysBlockDir, e.Name(), "loop", "backing_file")) //nolint:gosec // sysfs path
		if err != nil {
			continue // no backing file means the loop device is not in use
		}
		backing := strings.TrimSpace(string(data))
		if backing == "" {
			continue
		}
		devices = append(devices, loopDevice{name: e.Name(), backing: backing})
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].name < devices[j].name })
	return devices
}
