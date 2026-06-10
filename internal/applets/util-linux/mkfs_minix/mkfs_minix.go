// Package mkfsminix implements the mkfs.minix applet: create a Minix version-1
// filesystem on a device or image file.
package mkfsminix

import (
	"context"
	"encoding/binary"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mkfs.minix applet.
type Command struct{}

// New returns a mkfs.minix command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkfs.minix" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a Minix filesystem" }

const (
	blockSize    = 1024
	minixMagic   = 0x137F // version 1, 14-character file names
	inodesPerBlk = blockSize / 32
	bitsPerBlock = blockSize * 8
	sIFDIR       = 0o040000
	rootMode     = sIFDIR | 0o755
	dirEntrySize = 16         // 2-byte inode number + 14-byte name
	maxSize      = 0x10081C00 // Minix v1 maximum file size (7+512+512*512 zones)
	minBlocks    = 16         // refuse to format anything smaller
)

// Run executes mkfs.minix.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE [BLOCKS]", stdio.Err).WithHelp(command.Help{
		Description: "Create a Minix version-1 filesystem (1 KiB blocks, 14-character names) on DEVICE, " +
			"which may be a block device or an image file. BLOCKS sets the size in 1 KiB blocks; by " +
			"default the whole device/file is used. The filesystem is created with an empty root " +
			"directory.",
		Examples: []command.Example{
			{Command: "mkfs.minix disk.img", Explain: "Format the image as Minix."},
			{Command: "mkfs.minix /dev/sdb1 65536", Explain: "Format 64 MiB of the device."},
		},
		ExitStatus: "0  the filesystem was created.\n1  the device was missing or too small.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a device or image is required")
	}

	f, err := os.OpenFile(rest[0], os.O_RDWR, 0o644) //nolint:gosec // user-named device or image
	if err != nil {
		return command.Failuref("cannot open %s: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()

	nblocks, err := blockCount(f, rest)
	if err != nil {
		return command.Failuref("%v", err)
	}
	if nblocks < minBlocks {
		return command.Failuref("%s: device is too small (need at least %d blocks)", rest[0], minBlocks)
	}

	if err := writeMinix(f, nblocks); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// blockCount returns the number of 1 KiB blocks to format, from the optional
// BLOCKS operand or the file size.
func blockCount(f *os.File, args []string) (int, error) {
	if len(args) > 1 {
		n, err := parseUint(args[1])
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return int(info.Size() / blockSize), nil
}

// layout holds the computed on-disk geometry of a Minix v1 filesystem.
type layout struct {
	ninodes, nzones                     int
	imapBlocks, zmapBlocks, inodeBlocks int
	firstDataZone                       int
}

func upper(a, b int) int { return (a + b - 1) / b }

// computeLayout derives the bitmap/inode geometry for nblocks 1 KiB blocks.
func computeLayout(nblocks int) layout {
	nzones := nblocks
	ninodes := nblocks / 3
	if ninodes < 1 {
		ninodes = 1
	}
	ninodes = upper(ninodes, inodesPerBlk) * inodesPerBlk // fill the inode blocks

	imapBlocks := upper(ninodes+1, bitsPerBlock)
	inodeBlocks := upper(ninodes, inodesPerBlk)

	zmapBlocks := 0
	for {
		firstDataZone := 2 + imapBlocks + zmapBlocks + inodeBlocks
		next := upper(nzones-firstDataZone+1, bitsPerBlock)
		if next == zmapBlocks {
			break
		}
		zmapBlocks = next
	}
	return layout{
		ninodes:       ninodes,
		nzones:        nzones,
		imapBlocks:    imapBlocks,
		zmapBlocks:    zmapBlocks,
		inodeBlocks:   inodeBlocks,
		firstDataZone: 2 + imapBlocks + zmapBlocks + inodeBlocks,
	}
}

// writeMinix lays out the superblock, bitmaps, root inode, and root directory.
func writeMinix(f *os.File, nblocks int) error {
	l := computeLayout(nblocks)

	sb := make([]byte, blockSize)
	binary.LittleEndian.PutUint16(sb[0:], uint16(l.ninodes))
	binary.LittleEndian.PutUint16(sb[2:], uint16(l.nzones))
	binary.LittleEndian.PutUint16(sb[4:], uint16(l.imapBlocks))
	binary.LittleEndian.PutUint16(sb[6:], uint16(l.zmapBlocks))
	binary.LittleEndian.PutUint16(sb[8:], uint16(l.firstDataZone))
	binary.LittleEndian.PutUint16(sb[10:], 0) // log_zone_size
	binary.LittleEndian.PutUint32(sb[12:], maxSize)
	binary.LittleEndian.PutUint16(sb[16:], minixMagic)
	binary.LittleEndian.PutUint16(sb[18:], 1) // MINIX_VALID_FS
	if err := writeAt(f, sb, blockSize); err != nil {
		return err
	}

	// Inode bitmap: bit 0 is reserved, bit 1 is the root inode, and every bit
	// past the last inode is marked used so it is never allocated.
	imap := make([]byte, l.imapBlocks*blockSize)
	setBit(imap, 0)
	setBit(imap, 1)
	for b := l.ninodes + 1; b < l.imapBlocks*bitsPerBlock; b++ {
		setBit(imap, b)
	}
	if err := writeAt(f, imap, 2*blockSize); err != nil {
		return err
	}

	// Zone bitmap: bit 0 reserved, bit 1 is the root directory's data zone, and
	// every bit past the last zone is marked used.
	zmap := make([]byte, l.zmapBlocks*blockSize)
	setBit(zmap, 0)
	setBit(zmap, 1)
	validZones := l.nzones - l.firstDataZone + 1
	for b := validZones + 1; b < l.zmapBlocks*bitsPerBlock; b++ {
		setBit(zmap, b)
	}
	if err := writeAt(f, zmap, int64(2+l.imapBlocks)*blockSize); err != nil {
		return err
	}

	// Root inode (inode 1) sits at the start of the inode table.
	inodeTable := int64(2+l.imapBlocks+l.zmapBlocks) * blockSize
	ino := make([]byte, 32)
	binary.LittleEndian.PutUint16(ino[0:], rootMode)
	binary.LittleEndian.PutUint32(ino[4:], 2*dirEntrySize) // size: "." and ".."
	ino[13] = 2                                            // nlinks
	binary.LittleEndian.PutUint16(ino[14:], uint16(l.firstDataZone))
	if err := writeAt(f, ino, inodeTable); err != nil {
		return err
	}

	// Root directory data: "." and ".." both point at inode 1.
	dir := make([]byte, blockSize)
	binary.LittleEndian.PutUint16(dir[0:], 1)
	copy(dir[2:], ".")
	binary.LittleEndian.PutUint16(dir[dirEntrySize:], 1)
	copy(dir[dirEntrySize+2:], "..")
	return writeAt(f, dir, int64(l.firstDataZone)*blockSize)
}

func writeAt(f *os.File, data []byte, off int64) error {
	_, err := f.WriteAt(data, off)
	return err
}

// setBit sets bit n in the little-endian bitmap.
func setBit(bitmap []byte, n int) {
	bitmap[n/8] |= 1 << (uint(n) % 8)
}

// parseUint parses a positive base-10 integer.
func parseUint(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, command.Failuref("invalid block count: %q", s)
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, command.Failuref("invalid block count: %q", s)
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, nil
}
