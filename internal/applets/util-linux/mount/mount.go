// Package mount implements the mount applet: list the mounted filesystems.
// Mounting itself is privileged and is not performed by this slice.
package mount

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mount applet.
type Command struct{}

// New returns a mount command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mount" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List the mounted filesystems" }

// mountsPath is the kernel mount table; tests point it at a fixture.
var mountsPath = "/proc/mounts"

// Run executes mount.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t TYPE]", stdio.Err).WithHelp(command.Help{
		Description: "List the mounted filesystems, one per line as 'DEVICE on MOUNTPOINT type FSTYPE " +
			"(OPTIONS)', read from /proc/mounts. With -t, show only mounts of the given filesystem " +
			"type. Performing a mount is privileged and is not done by this build.",
		Examples: []command.Example{
			{Command: "mount", Explain: "List every mounted filesystem."},
			{Command: "mount -t ext4", Explain: "List only ext4 mounts."},
		},
		ExitStatus: "0  the table was listed.\n1  the mount table could not be read, or a mount was requested.",
	})
	typeFilter := fs.StringP("types", "t", "", "only show mounts of this filesystem type")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if len(fs.Args()) > 0 {
		_, _ = fmt.Fprintln(stdio.Err, "mount: performing a mount is not supported by this build; "+
			"run with no operands to list the mounted filesystems")
		return command.SilentFailure()
	}

	mounts, err := readMounts()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "mount: %v\n", err)
		return command.SilentFailure()
	}
	for _, m := range mounts {
		if *typeFilter != "" && m.fstype != *typeFilter {
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s on %s type %s (%s)\n", m.device, m.mountpoint, m.fstype, m.options)
	}
	return nil
}

type mountEntry struct {
	device, mountpoint, fstype, options string
}

// readMounts parses the kernel mount table.
func readMounts() ([]mountEntry, error) {
	f, err := os.Open(mountsPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var entries []mountEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 4 {
			continue
		}
		entries = append(entries, mountEntry{
			device:     unescape(fields[0]),
			mountpoint: unescape(fields[1]),
			fstype:     fields[2],
			options:    fields[3],
		})
	}
	return entries, sc.Err()
}

// unescape decodes the octal escapes (\040 space, \011 tab, …) used in
// /proc/mounts device and mount-point fields.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+4 <= len(s) && isOctal(s[i+1:i+4]) {
			b.WriteByte(octalByte(s[i+1 : i+4]))
			i += 3
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isOctal(s string) bool {
	if len(s) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		if s[i] < '0' || s[i] > '7' {
			return false
		}
	}
	return true
}

func octalByte(s string) byte {
	return byte((s[0]-'0')<<6 | (s[1]-'0')<<3 | (s[2] - '0'))
}
