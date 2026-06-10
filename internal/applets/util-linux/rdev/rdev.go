// Package rdev implements the rdev applet: print the device of the root
// filesystem.
package rdev

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rdev applet.
type Command struct{}

// New returns an rdev command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rdev" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the root filesystem device" }

// procMounts is the mount table; tests point it at a fixture.
var procMounts = "/proc/mounts"

// Run executes rdev.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the device mounted as the root filesystem (/), in the historical " +
			"'DEVICE /' form. Setting the root device, which old rdev did by patching a kernel " +
			"image, is not supported.",
		Examples: []command.Example{
			{Command: "rdev", Explain: "Print the root device, e.g. '/dev/sda1 /'."},
		},
		ExitStatus: "0  the root device was found.\n1  it could not be determined.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	device, err := rootDevice(procMounts)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "rdev: %v\n", err)
		return command.SilentFailure()
	}
	_, _ = fmt.Fprintf(stdio.Out, "%s /\n", device)
	return nil
}

// rootDevice returns the device mounted at "/" from the mount table.
func rootDevice(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // the mount table path
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) >= 2 && fields[1] == "/" {
			return fields[0], nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("could not determine the root device")
}
