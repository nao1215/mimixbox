// Package mkfsvfat implements the mkfs.vfat applet (also mkdosfs): create a
// FAT16 filesystem on a device or image file.
package mkfsvfat

import (
	"context"
	"encoding/binary"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mkfs.vfat applet. It is also registered under the traditional
// name mkdosfs.
type Command struct{ name string }

// New returns a mkfs.vfat command.
func New() *Command { return &Command{name: "mkfs.vfat"} }

// NewMkdosfs returns the same applet under the name mkdosfs.
func NewMkdosfs() *Command { return &Command{name: "mkdosfs"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a FAT16 filesystem" }

const (
	sectorSize      = 512
	rootEntries     = 512
	reservedSectors = 1
	numFATs         = 2
	mediaByte       = 0xF8
	// FAT16 is only valid for at least this many data clusters.
	minFAT16Clusters = 4085
)

// Run executes mkfs.vfat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n LABEL] DEVICE", stdio.Err).WithHelp(command.Help{
		Description: "Create a FAT16 filesystem on DEVICE (a block device or image file) with an empty " +
			"root directory. -n sets an 11-character volume label. The device must be large enough for " +
			"a FAT16 filesystem (a few MiB); smaller sizes are refused rather than silently making FAT12.",
		Examples: []command.Example{
			{Command: "mkfs.vfat disk.img", Explain: "Format the image as FAT16."},
			{Command: "mkfs.vfat -n DATA disk.img", Explain: "Format it with the label DATA."},
		},
		ExitStatus: "0  the filesystem was created.\n1  the device was missing or too small.",
	})
	label := fs.StringP("label", "n", "NO NAME", "volume label (up to 11 characters)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	hasLabel := fs.Changed("label")

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device or image is required")
	}

	f, err := os.OpenFile(rest[0], os.O_RDWR, 0o644) //nolint:gosec // user-named device or image
	if err != nil {
		return command.Failuref("cannot open %s: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return command.Failuref("cannot stat %s: %v", rest[0], err)
	}
	totalSectors := int(info.Size() / sectorSize)

	if err := writeFAT16(f, totalSectors, *label, hasLabel); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// geometry holds the computed FAT16 layout.
type geometry struct {
	totalSectors   int
	rootDirSectors int
	sectorsPerFAT  int
	clusters       int
}

func upper(a, b int) int { return (a + b - 1) / b }

// computeGeometry derives the FAT16 layout for totalSectors 512-byte sectors,
// using one sector per cluster.
func computeGeometry(totalSectors int) geometry {
	rootDirSectors := upper(rootEntries*32, sectorSize)
	sectorsPerFAT := 1
	for {
		dataSectors := totalSectors - reservedSectors - rootDirSectors - numFATs*sectorsPerFAT
		clusters := dataSectors // one sector per cluster
		next := upper((clusters+2)*2, sectorSize)
		if next == sectorsPerFAT {
			return geometry{totalSectors, rootDirSectors, sectorsPerFAT, clusters}
		}
		sectorsPerFAT = next
	}
}

// writeFAT16 writes the boot sector, the two FATs, and the root directory. When
// hasLabel is set, the root directory gets a matching volume-label entry.
func writeFAT16(f *os.File, totalSectors int, label string, hasLabel bool) error {
	g := computeGeometry(totalSectors)
	if g.clusters < minFAT16Clusters {
		return command.Failuref("device is too small for FAT16 (%d clusters, need %d)",
			g.clusters, minFAT16Clusters)
	}

	boot := make([]byte, sectorSize)
	boot[0], boot[1], boot[2] = 0xEB, 0x3C, 0x90 // jump
	copy(boot[3:11], "MIMIXBOX")
	binary.LittleEndian.PutUint16(boot[11:], sectorSize)
	boot[13] = 1 // sectors per cluster
	binary.LittleEndian.PutUint16(boot[14:], reservedSectors)
	boot[16] = numFATs
	binary.LittleEndian.PutUint16(boot[17:], rootEntries)
	if totalSectors < 0x10000 {
		binary.LittleEndian.PutUint16(boot[19:], uint16(totalSectors))
	}
	boot[21] = mediaByte
	binary.LittleEndian.PutUint16(boot[22:], uint16(g.sectorsPerFAT))
	binary.LittleEndian.PutUint16(boot[24:], 32) // sectors per track
	binary.LittleEndian.PutUint16(boot[26:], 64) // heads
	if totalSectors >= 0x10000 {
		binary.LittleEndian.PutUint32(boot[32:], uint32(totalSectors))
	}
	boot[36] = 0x80 // drive number
	boot[38] = 0x29 // extended boot signature
	binary.LittleEndian.PutUint32(boot[39:], 0x12345678)
	copy(boot[43:54], padLabel(label))
	copy(boot[54:62], "FAT16   ")
	boot[510], boot[511] = 0x55, 0xAA
	if _, err := f.WriteAt(boot, 0); err != nil {
		return err
	}

	// Each FAT starts with the media byte and an end-of-chain marker.
	fat := make([]byte, g.sectorsPerFAT*sectorSize)
	fat[0], fat[1], fat[2], fat[3] = mediaByte, 0xFF, 0xFF, 0xFF // entry 0 and 1
	for i := 0; i < numFATs; i++ {
		off := int64(reservedSectors+i*g.sectorsPerFAT) * sectorSize
		if _, err := f.WriteAt(fat, off); err != nil {
			return err
		}
	}

	// The root directory is empty, except for a volume-label entry when -n was
	// given (so a checker finds the label both in the boot sector and the root).
	root := make([]byte, g.rootDirSectors*sectorSize)
	if hasLabel {
		copy(root[0:11], padLabel(label))
		root[11] = 0x08 // ATTR_VOLUME_ID
	}
	rootOff := int64(reservedSectors+numFATs*g.sectorsPerFAT) * sectorSize
	_, err := f.WriteAt(root, rootOff)
	return err
}

// padLabel returns the 11-byte, space-padded, uppercase-ish volume label.
func padLabel(label string) []byte {
	out := []byte("           ") // 11 spaces
	copy(out, label)
	if len(label) > 11 {
		out = out[:11]
	}
	return out
}
