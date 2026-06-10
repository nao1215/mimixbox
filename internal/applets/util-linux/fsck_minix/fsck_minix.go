// Package fsckminix implements the fsck.minix applet: check a Minix filesystem
// by validating its superblock and reporting its geometry.
package fsckminix

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fsck.minix applet.
type Command struct{}

// New returns a fsck.minix command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fsck.minix" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Check a Minix filesystem" }

const (
	blockSize        = 1024
	superblockOffset = 1024
)

// magics maps each Minix superblock magic to a human description.
var magics = map[uint16]string{
	0x137F: "Minix v1, 14-character names",
	0x138F: "Minix v1, 30-character names",
	0x2468: "Minix v2, 14-character names",
	0x2478: "Minix v2, 30-character names",
}

// Run executes fsck.minix.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Check the Minix filesystem on DEVICE (a block device or image file) by validating " +
			"its superblock and reporting its version and geometry: the inode count, zone count, and " +
			"first data zone. -f is accepted for compatibility. This build performs the superblock " +
			"check, not a full inode/zone walk.",
		Examples: []command.Example{
			{Command: "fsck.minix disk.img", Explain: "Check the Minix filesystem in the image."},
		},
		ExitStatus: "0  the superblock is a valid Minix filesystem.\n1  the device was unreadable or not Minix.",
	})
	_ = fs.BoolP("force", "f", false, "force a check (accepted for compatibility)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device or image is required")
	}

	sb, err := readSuperblock(rest[0])
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}

	desc, ok := magics[sb.magic]
	if !ok {
		return command.Failuref("%s: bad magic number in super-block (%#04x); not a Minix filesystem",
			rest[0], sb.magic)
	}

	_, _ = fmt.Fprintf(stdio.Out, "%s: %s\n", rest[0], desc)
	_, _ = fmt.Fprintf(stdio.Out, "%d inodes\n", sb.ninodes)
	_, _ = fmt.Fprintf(stdio.Out, "%d zones\n", sb.nzones)
	_, _ = fmt.Fprintf(stdio.Out, "firstdatazone=%d\n", sb.firstZone)
	return nil
}

type superblock struct {
	ninodes, nzones, firstZone int
	magic                      uint16
}

// readSuperblock reads and parses the Minix superblock at offset 1024.
func readSuperblock(path string) (*superblock, error) {
	f, err := os.Open(path) //nolint:gosec // user-named device or image
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, blockSize)
	if _, err := f.ReadAt(buf, superblockOffset); err != nil {
		return nil, fmt.Errorf("cannot read super-block: %w", err)
	}
	return &superblock{
		ninodes:   int(binary.LittleEndian.Uint16(buf[0:])),
		nzones:    int(binary.LittleEndian.Uint16(buf[2:])),
		firstZone: int(binary.LittleEndian.Uint16(buf[8:])),
		magic:     binary.LittleEndian.Uint16(buf[16:]),
	}, nil
}
