// Package fsck implements the fsck applet: detect the filesystem type on a
// device or image and report it. A full consistency walk is delegated to the
// type-specific checkers (e.g. fsck.minix).
package fsck

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fsck applet.
type Command struct{}

// New returns a fsck command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fsck" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Detect and report a filesystem type" }

// minixMagics are the Minix superblock magics.
var minixMagics = map[uint16]bool{0x137F: true, 0x138F: true, 0x2468: true, 0x2478: true}

// Run executes fsck.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Detect the filesystem on DEVICE (a block device or image file) by inspecting its " +
			"on-disk signatures and report the type (ext2/3/4, minix, vfat, or swap). A full " +
			"consistency check is delegated to the type-specific checkers such as fsck.minix.",
		Examples: []command.Example{
			{Command: "fsck disk.img", Explain: "Report the filesystem type of the image."},
		},
		ExitStatus: "0  a known filesystem was detected.\n1  the device was unreadable or unrecognized.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device or image is required")
	}

	buf, err := readHead(rest[0])
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}

	fstype := detect(buf)
	if fstype == "" {
		return command.Failuref("%s: unable to detect a known filesystem type", rest[0])
	}
	_, _ = fmt.Fprintf(stdio.Out, "%s: %s\n", rest[0], fstype)
	return nil
}

// detect identifies the filesystem from the on-disk signatures in buf.
func detect(buf []byte) string {
	// ext2/3/4: s_magic 0xEF53 at superblock offset 56 (block 1).
	if len(buf) >= 1024+58 && binary.LittleEndian.Uint16(buf[1024+56:]) == 0xEF53 {
		return "ext2/ext3/ext4"
	}
	// minix: s_magic at superblock offset 16.
	if len(buf) >= 1024+18 && minixMagics[binary.LittleEndian.Uint16(buf[1024+16:])] {
		return "minix"
	}
	// vfat: the 0x55AA boot signature plus a FATxx type label.
	if len(buf) >= 512 && buf[510] == 0x55 && buf[511] == 0xAA {
		if hasFATLabel(buf, 54) || hasFATLabel(buf, 82) {
			return "vfat"
		}
	}
	// swap: the "SWAPSPACE2" signature at the end of the first 4 KiB page.
	if len(buf) >= 4096 && string(buf[4086:4096]) == "SWAPSPACE2" {
		return "swap"
	}
	return ""
}

// hasFATLabel reports whether buf holds an "FAT" type label at off.
func hasFATLabel(buf []byte, off int) bool {
	return len(buf) >= off+3 && string(buf[off:off+3]) == "FAT"
}

// readHead reads the first 8 KiB of path (enough for every signature checked).
func readHead(path string) ([]byte, error) {
	f, err := os.Open(path) //nolint:gosec // user-named device or image
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 8192)
	n, err := f.ReadAt(buf, 0)
	if n == 0 && err != nil {
		return nil, fmt.Errorf("cannot read: %w", err)
	}
	return buf[:n], nil
}
