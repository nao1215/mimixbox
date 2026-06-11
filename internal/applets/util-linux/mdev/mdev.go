// Package mdev implements the mdev applet in scan mode: create device nodes in
// /dev from the entries advertised under /sys/class.
package mdev

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the mdev applet.
type Command struct{}

// New returns a mdev command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mdev" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create /dev nodes from /sys (scan mode)" }

// Injected so the scan and the privileged mknod are testable.
var (
	sysClassDir = "/sys/class"
	devDir      = "/dev"
	mknodFn     = func(path string, isBlock bool, major, minor uint32) error {
		mode := uint32(unix.S_IFCHR | 0o660)
		if isBlock {
			mode = unix.S_IFBLK | 0o660
		}
		return unix.Mknod(path, mode, int(unix.Mkdev(major, minor)))
	}
)

// Run executes mdev.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-s", stdio.Err).WithHelp(command.Help{
		Description: "In scan mode (-s), walk /sys/class and create a device node under /dev for every " +
			"device that advertises a 'dev' (major:minor) attribute, as a block node for the block " +
			"class and a character node otherwise. The hotplug-helper mode is not implemented. " +
			"Creating device nodes requires privilege.",
		Examples: []command.Example{
			{Command: "mdev -s", Explain: "Populate /dev from /sys."},
		},
		ExitStatus: "0  the nodes were created.\n1  -s was not given or a node could not be created.",
	})
	scan := fs.BoolP("scan", "s", false, "scan /sys and create device nodes")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*scan {
		_, _ = fmt.Fprintln(stdio.Err, "mdev: only scan mode (-s) is supported by this build")
		return command.SilentFailure()
	}

	created, failed := 0, false
	for _, d := range scanDevices() {
		path := filepath.Join(devDir, d.name)
		if err := mknodFn(path, d.isBlock, d.major, d.minor); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mdev: %s: %v\n", path, err)
			failed = true
			continue
		}
		created++
	}
	_, _ = fmt.Fprintf(stdio.Out, "mdev: created %d device node(s)\n", created)
	if failed {
		return command.SilentFailure()
	}
	return nil
}

type device struct {
	name         string
	isBlock      bool
	major, minor uint32
}

// scanDevices returns every device advertised under /sys/class that has a valid
// dev attribute.
func scanDevices() []device {
	classes, err := os.ReadDir(sysClassDir)
	if err != nil {
		return nil
	}
	var devices []device
	for _, class := range classes {
		isBlock := class.Name() == "block"
		classPath := filepath.Join(sysClassDir, class.Name())
		entries, err := os.ReadDir(classPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			major, minor, ok := readDev(filepath.Join(classPath, e.Name(), "dev"))
			if !ok {
				continue
			}
			devices = append(devices, device{name: e.Name(), isBlock: isBlock, major: major, minor: minor})
		}
	}
	return devices
}

// readDev parses a "major:minor" sysfs dev attribute file.
func readDev(path string) (major, minor uint32, ok bool) {
	data, err := os.ReadFile(path) //nolint:gosec // sysfs attribute path
	if err != nil {
		return 0, 0, false
	}
	maj, min, found := strings.Cut(strings.TrimSpace(string(data)), ":")
	if !found {
		return 0, 0, false
	}
	majN, err1 := strconv.ParseUint(maj, 10, 32)
	minN, err2 := strconv.ParseUint(min, 10, 32)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return uint32(majN), uint32(minN), true
}
