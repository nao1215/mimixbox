// Package lsscsi implements the lsscsi applet: list the SCSI (and SCSI-emulated)
// devices reported by the kernel under /sys/bus/scsi/devices.
package lsscsi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the lsscsi applet.
type Command struct{}

// New returns an lsscsi command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsscsi" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List SCSI devices" }

// scsiDevices is the sysfs SCSI directory; tests point it at a fixture.
var scsiDevices = "/sys/bus/scsi/devices"

// device is one parsed SCSI device entry.
type device struct {
	hctl   string // host:channel:target:lun, e.g. "0:0:0:0"
	typ    string // numeric type code from sysfs
	vendor string
	model  string
	rev    string
	node   string // /dev node, e.g. /dev/sda, when discoverable
}

// scsiTypes maps the numeric sysfs type code to lsscsi's short type name.
var scsiTypes = map[string]string{
	"0":  "disk",
	"1":  "tape",
	"2":  "printer",
	"3":  "process",
	"4":  "worm",
	"5":  "cd/dvd",
	"6":  "scanner",
	"7":  "optical",
	"8":  "mediumx",
	"9":  "comms",
	"12": "raid",
	"13": "enclosu",
	"14": "rbc",
	"17": "osd",
}

// Run executes lsscsi.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c] [-d]", stdio.Err).WithHelp(command.Help{
		Description: "List SCSI devices (and devices the kernel presents through the SCSI subsystem, such as " +
			"USB storage and SATA disks) read from /sys/bus/scsi/devices. Each line shows the [H:C:T:L] " +
			"address, the device type, the vendor/model/revision strings, and the /dev node when one can be " +
			"resolved. This is a read-only command; it never touches the devices themselves.",
		Examples: []command.Example{
			{Command: "lsscsi", Explain: "List all SCSI devices."},
			{Command: "lsscsi -c", Explain: "List devices with the type spelled out in full."},
		},
		ExitStatus: "0  the device list was produced (even if empty).\n1  the sysfs SCSI tree is unreadable.",
		Notes: []string{
			"Devices appear only if the running kernel exports /sys/bus/scsi/devices.",
		},
	})
	classic := fs.BoolP("classic", "c", false, "render the device type using its full name")
	_ = fs.BoolP("device", "d", false, "show the major:minor of each device node (accepted, sysfs based)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	devs, err := scan(scsiDevices)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "lsscsi: %s\n", command.FileError(scsiDevices, err))
		return command.SilentFailure()
	}
	for _, d := range devs {
		_, _ = fmt.Fprintln(stdio.Out, format(d, *classic))
	}
	return nil
}

// scan reads root and returns the SCSI devices found there, sorted by address.
func scan(root string) ([]device, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var devs []device
	for _, e := range entries {
		name := e.Name()
		if !isHCTL(name) {
			continue
		}
		devs = append(devs, readDevice(filepath.Join(root, name), name))
	}
	sort.Slice(devs, func(i, j int) bool { return lessHCTL(devs[i].hctl, devs[j].hctl) })
	return devs, nil
}

// readDevice reads the sysfs attributes of a single device directory.
func readDevice(dir, hctl string) device {
	d := device{
		hctl:   hctl,
		typ:    attr(dir, "type"),
		vendor: attr(dir, "vendor"),
		model:  attr(dir, "model"),
		rev:    attr(dir, "rev"),
	}
	d.node = blockNode(dir)
	return d
}

// blockNode resolves the /dev node for a device by reading its block/ subdir.
func blockNode(dir string) string {
	entries, err := os.ReadDir(filepath.Join(dir, "block"))
	if err != nil || len(entries) == 0 {
		return ""
	}
	return "/dev/" + entries[0].Name()
}

// attr reads a one-line sysfs attribute, trimmed.
func attr(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // a /sys attribute path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// format renders one device the way lsscsi does.
func format(d device, classic bool) string {
	typeName := typeName(d.typ)
	if classic {
		typeName = fmt.Sprintf("(0x%s)", d.typ)
	}
	node := d.node
	if node == "" {
		node = "-"
	}
	return fmt.Sprintf("[%s]    %-7s %-8s %-16s %-4s  %s",
		d.hctl, typeName, d.vendor, d.model, d.rev, node)
}

// typeName maps a numeric type to its short label, falling back to the number.
func typeName(code string) string {
	if name, ok := scsiTypes[code]; ok {
		return name
	}
	if code == "" {
		return "-"
	}
	return "(0x" + code + ")"
}

// isHCTL reports whether name is an H:C:T:L address like "6:0:0:0".
func isHCTL(name string) bool {
	parts := strings.Split(name, ":")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

// lessHCTL orders two H:C:T:L addresses numerically, field by field.
func lessHCTL(a, b string) bool {
	ap := strings.Split(a, ":")
	bp := strings.Split(b, ":")
	for i := 0; i < 4; i++ {
		ai, _ := strconv.Atoi(ap[i])
		bi, _ := strconv.Atoi(bp[i])
		if ai != bi {
			return ai < bi
		}
	}
	return false
}
