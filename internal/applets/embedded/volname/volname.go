// Package volname implements the volname applet: print the volume name of an
// ISO 9660 filesystem (typically a CD-ROM device or image).
package volname

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the volname applet.
type Command struct{}

// New returns a volname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "volname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the volume name of an ISO 9660 filesystem" }

// ISO 9660 stores the 32-byte volume identifier in the Primary Volume
// Descriptor, which starts at logical sector 16 (2048-byte sectors). The
// identifier field is at offset 40 within that descriptor.
const (
	pvdOffset       = 16 * 2048
	volIDFieldOff   = 40
	volIDFieldBytes = 32
)

// defaultDevice is read when no operand is given, matching the traditional
// volname behavior of defaulting to the primary CD-ROM device.
var defaultDevice = "/dev/cdrom"

// Run executes volname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[DEVICE]", stdio.Err).WithHelp(command.Help{
		Description: "Print the 32-character volume identifier from the ISO 9660 Primary Volume Descriptor of " +
			"DEVICE (a block device or image file). With no operand /dev/cdrom is read. This is a read-only " +
			"command; it never writes to the device.",
		Examples: []command.Example{
			{Command: "volname /dev/sr0", Explain: "Print the label of the disc in /dev/sr0."},
			{Command: "volname disc.iso", Explain: "Print the label of an ISO image file."},
		},
		ExitStatus: "0  a label was read.\n1  the device could not be read or has no ISO 9660 descriptor.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	device := defaultDevice
	if rest := fs.Args(); len(rest) > 0 {
		if len(rest) > 1 {
			return command.Failuref("at most one device operand is allowed")
		}
		device = rest[0]
	}

	label, err := readVolumeName(device)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "volname: %s\n", command.FileError(device, err))
		return command.SilentFailure()
	}
	_, _ = fmt.Fprintln(stdio.Out, label)
	return nil
}

// readVolumeName reads and trims the ISO 9660 volume identifier from device.
func readVolumeName(device string) (string, error) {
	f, err := os.Open(device) //nolint:gosec // user-named device/image
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, volIDFieldBytes)
	if _, err := f.ReadAt(buf, pvdOffset+volIDFieldOff); err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimRight(string(buf), " \x00"), nil
}
