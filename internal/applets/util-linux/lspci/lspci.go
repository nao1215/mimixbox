// Package lspci implements the lspci applet: list PCI devices read from
// /sys/bus/pci/devices, in the numeric class/vendor:device form.
package lspci

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the lspci applet.
type Command struct{}

// New returns an lspci command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lspci" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List PCI devices" }

// pciDevices is the sysfs PCI directory; tests point it at a fixture.
var pciDevices = "/sys/bus/pci/devices"

// Run executes lspci.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n]", stdio.Err).WithHelp(command.Help{
		Description: "List the PCI devices under /sys/bus/pci/devices as 'SLOT CLASS: VENDOR:DEVICE', " +
			"with a revision when non-zero. Vendor/device names from a pci.ids database are not " +
			"resolved, so the output is always numeric.",
		Examples: []command.Example{
			{Command: "lspci", Explain: "List the PCI devices."},
		},
	})
	_ = fs.BoolP("numeric", "n", false, "show numeric IDs (always on in this implementation)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	entries, err := os.ReadDir(pciDevices)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "lspci: %s\n", command.FileError(pciDevices, err))
		return command.SilentFailure()
	}

	var slots []string
	for _, e := range entries {
		slots = append(slots, e.Name())
	}
	sort.Strings(slots)

	for _, slot := range slots {
		c.printDevice(stdio.Out, slot)
	}
	return nil
}

func (c *Command) printDevice(out io.Writer, slot string) {
	dir := filepath.Join(pciDevices, slot)
	class := hexField(dir, "class")
	vendor := hexField(dir, "vendor")
	device := hexField(dir, "device")
	rev := hexField(dir, "revision")

	// The class field is 6 hex digits (class, subclass, prog-if); lspci shows
	// the first four (class + subclass).
	if len(class) > 4 {
		class = class[:4]
	}

	line := fmt.Sprintf("%s %s: %s:%s", displaySlot(slot), class, vendor, device)
	if rev != "" && rev != "00" {
		line += fmt.Sprintf(" (rev %s)", rev)
	}
	_, _ = fmt.Fprintln(out, line)
}

// displaySlot drops the default "0000:" PCI domain, as lspci does.
func displaySlot(slot string) string {
	return strings.TrimPrefix(slot, "0000:")
}

// hexField reads a "0x..." sysfs attribute and returns its digits without the
// prefix.
func hexField(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a /sys attribute path
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(data)), "0x")
}
