// Package lsusb implements the lsusb applet: list USB devices read from
// /sys/bus/usb/devices.
package lsusb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the lsusb applet.
type Command struct{}

// New returns an lsusb command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsusb" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List USB devices" }

// usbDevices is the sysfs USB directory; tests point it at a fixture.
var usbDevices = "/sys/bus/usb/devices"

type usbDevice struct {
	bus    int
	dev    int
	vendor string
	prod   string
	name   string
}

// Run executes lsusb.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "List the USB devices under /sys/bus/usb/devices as 'Bus BBB Device DDD: ID " +
			"vendor:product NAME', sorted by bus and device number. The name is taken from the " +
			"device's own manufacturer/product strings (no usb.ids database lookup).",
		Examples: []command.Example{
			{Command: "lsusb", Explain: "List the USB devices."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	entries, err := os.ReadDir(usbDevices)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "lsusb: %s\n", command.FileError(usbDevices, err))
		return command.SilentFailure()
	}

	var devices []usbDevice
	for _, e := range entries {
		dir := filepath.Join(usbDevices, e.Name())
		vendor := field(dir, "idVendor")
		if vendor == "" {
			continue // interfaces and other non-device nodes have no idVendor
		}
		devices = append(devices, usbDevice{
			bus:    atoi(field(dir, "busnum")),
			dev:    atoi(field(dir, "devnum")),
			vendor: vendor,
			prod:   field(dir, "idProduct"),
			name:   strings.TrimSpace(field(dir, "manufacturer") + " " + field(dir, "product")),
		})
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].bus != devices[j].bus {
			return devices[i].bus < devices[j].bus
		}
		return devices[i].dev < devices[j].dev
	})

	for _, d := range devices {
		printDevice(stdio.Out, d)
	}
	return nil
}

func printDevice(out io.Writer, d usbDevice) {
	line := fmt.Sprintf("Bus %03d Device %03d: ID %s:%s", d.bus, d.dev, d.vendor, d.prod)
	if d.name != "" {
		line += " " + d.name
	}
	_, _ = fmt.Fprintln(out, line)
}

func field(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a /sys attribute path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
