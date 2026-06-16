// Package mkfsvfat implements the mkfs.vfat applet (also mkdosfs): create a
// FAT16 filesystem on a device or image file.
package mkfsvfat

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"unsafe"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Front-end command names. mkfs.vfat is canonical; mkdosfs is the traditional
// alias. Both drive the same FAT16 builder.
const (
	cmdMkfsVfat = "mkfs.vfat"
	cmdMkdosfs  = "mkdosfs"
)

// aliasConfig is the metadata for one front-end name under which the FAT16
// builder is exposed. Every alias shares the builder logic and differs only by
// the command name it reports.
type aliasConfig struct {
	synopsis string
}

// aliases is the name -> config table that drives the FAT16 front-ends.
var aliases = map[string]aliasConfig{
	cmdMkfsVfat: {synopsis: "Create a FAT16 filesystem"},
	cmdMkdosfs:  {synopsis: "Create a FAT16 filesystem"},
}

// Command is the mkfs.vfat applet. It is also registered under the traditional
// name mkdosfs.
type Command struct{ name string }

// New returns a mkfs.vfat command.
func New() *Command { return &Command{name: cmdMkfsVfat} }

// NewMkdosfs returns the same applet under the name mkdosfs.
func NewMkdosfs() *Command { return &Command{name: cmdMkdosfs} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return aliases[c.name].synopsis }

const (
	sectorSize      = 512
	rootEntries     = 512
	reservedSectors = 1
	numFATs         = 2
	mediaByte       = 0xF8
	// FAT16 is valid only for a data-cluster count in this inclusive range.
	minFAT16Clusters = 4085
	maxFAT16Clusters = 65524
)

// clusterSizes are the sectors-per-cluster values tried, smallest first, so the
// cluster count lands in the FAT16 range as the device grows.
var clusterSizes = []int{1, 2, 4, 8, 16, 32, 64}

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

	totalSectors, err := deviceSectors(f)
	if err != nil {
		return command.Failuref("cannot size %s: %v", rest[0], err)
	}

	if err := writeFAT16(f, totalSectors, *label, hasLabel); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// deviceSectors returns the size of f in 512-byte sectors. For a regular file
// the stat size is used; for a block device (where st_size is 0) the real
// capacity is queried with BLKGETSIZE64.
func deviceSectors(f *os.File) (int, error) {
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	size := info.Size()
	if size == 0 {
		var bytes uint64
		_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uintptr(unix.BLKGETSIZE64), uintptr(unsafe.Pointer(&bytes)))
		if errno != 0 {
			return 0, errno
		}
		size = int64(bytes)
	}
	return int(size / sectorSize), nil
}

// geometry holds the computed FAT16 layout.
type geometry struct {
	totalSectors      int
	rootDirSectors    int
	sectorsPerFAT     int
	clusters          int
	sectorsPerCluster int
}

func upper(a, b int) int { return (a + b - 1) / b }

// computeGeometry derives a FAT16 layout for totalSectors 512-byte sectors,
// choosing the smallest sectors-per-cluster that keeps the data-cluster count
// within the valid FAT16 range. It errors if no cluster size fits (the device
// is too small for FAT16, or large enough to need FAT32).
func computeGeometry(totalSectors int) (geometry, error) {
	rootDirSectors := upper(rootEntries*32, sectorSize)
	tooSmall := false
	for _, spc := range clusterSizes {
		sectorsPerFAT := 1
		var clusters int
		for {
			dataSectors := totalSectors - reservedSectors - rootDirSectors - numFATs*sectorsPerFAT
			clusters = dataSectors / spc
			next := upper((clusters+2)*2, sectorSize)
			if next == sectorsPerFAT {
				break
			}
			sectorsPerFAT = next
		}
		if clusters < minFAT16Clusters {
			tooSmall = true
			continue
		}
		if clusters <= maxFAT16Clusters {
			return geometry{totalSectors, rootDirSectors, sectorsPerFAT, clusters, spc}, nil
		}
	}
	if tooSmall {
		return geometry{}, errors.New("device is too small for FAT16")
	}
	return geometry{}, errors.New("device is too large for FAT16; use FAT32")
}

// writeFAT16 writes the boot sector, the two FATs, and the root directory. When
// hasLabel is set, the root directory gets a matching volume-label entry.
func writeFAT16(f *os.File, totalSectors int, label string, hasLabel bool) error {
	g, err := computeGeometry(totalSectors)
	if err != nil {
		return err
	}

	boot := make([]byte, sectorSize)
	boot[0], boot[1], boot[2] = 0xEB, 0x3C, 0x90 // jump
	copy(boot[3:11], "MIMIXBOX")
	binary.LittleEndian.PutUint16(boot[11:], sectorSize)
	boot[13] = byte(g.sectorsPerCluster)
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
	_, err = f.WriteAt(root, rootOff)
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
