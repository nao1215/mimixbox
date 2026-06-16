// Package mke2fs implements the mke2fs applet (also mkfs.ext2): create a small
// single-block-group ext2 filesystem on a device or image file.
package mke2fs

import (
	"context"
	"encoding/binary"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Front-end command names. mke2fs is canonical; mkfs.ext2 is the traditional
// mkfs-style alias. Both drive the same ext2 builder.
const (
	cmdMke2fs   = "mke2fs"
	cmdMkfsExt2 = "mkfs.ext2"
)

// aliasConfig is the metadata for one front-end name under which the ext2
// builder is exposed. Every alias shares the builder logic and differs only by
// the command name it reports.
type aliasConfig struct {
	synopsis string
}

// aliases is the name -> config table that drives the ext2 front-ends.
var aliases = map[string]aliasConfig{
	cmdMke2fs:   {synopsis: "Create an ext2 filesystem"},
	cmdMkfsExt2: {synopsis: "Create an ext2 filesystem"},
}

// Command is the mke2fs applet. It is also registered as mkfs.ext2.
type Command struct{ name string }

// New returns a mke2fs command.
func New() *Command { return &Command{name: cmdMke2fs} }

// NewMkfsExt2 returns the same applet under the name mkfs.ext2.
func NewMkfsExt2() *Command { return &Command{name: cmdMkfsExt2} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return aliases[c.name].synopsis }

// now is indirected so the recorded timestamps are deterministic in tests.
var now = time.Now

const (
	blockSize            = 1024
	inodeSize            = 128
	firstDataBlock       = 1
	blocksPerGroup       = 8192
	inodesPerBlock       = blockSize / inodeSize // 8
	inodeRatio           = 16384
	extMagic             = 0xEF53
	rootInode            = 2
	lostFoundInode       = 11
	firstInode           = 11
	featIncompatFiletype = 0x0002
	sIFDIR               = 0x4000
)

// Run executes mke2fs.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DEVICE [BLOCKS]", stdio.Err).WithHelp(command.Help{
		Description: "Create an ext2 filesystem (1 KiB blocks, 128-byte inodes, a single block group) " +
			"on DEVICE, a block device or image file, with a root directory and a lost+found " +
			"directory. BLOCKS sets the size in 1 KiB blocks; by default the whole device/file is " +
			"used. The size must fit one block group (up to 8192 blocks).",
		Examples: []command.Example{
			{Command: "mke2fs disk.img", Explain: "Format the image as ext2."},
			{Command: "mkfs.ext2 -F disk.img 4096", Explain: "Format 4 MiB as ext2."},
		},
		ExitStatus: "0  the filesystem was created.\n1  the device was missing, too small, or too large.",
	})
	_ = fs.BoolP("force", "F", false, "force creation (accepted for compatibility)")

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
	if err := writeExt2(f, nblocks); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

func blockCount(f *os.File, args []string) (int, error) {
	if len(args) > 1 {
		return parseUint(args[1])
	}
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return int(info.Size() / blockSize), nil
}

// layout is the computed single-group ext2 geometry.
type layout struct {
	blocks           int
	inodes           int
	inodeTableBlocks int
	blockBitmapBlock int
	inodeBitmapBlock int
	inodeTableBlock  int
	rootDirBlock     int
	lostFoundBlock   int
	usedBlocks       int
	usedInodes       int
}

func roundUp(a, b int) int { return ((a + b - 1) / b) * b }

// computeLayout derives the single-group layout for nblocks 1 KiB blocks.
func computeLayout(nblocks int) (layout, error) {
	if nblocks < 64 {
		return layout{}, command.Failuref("filesystem too small (need at least 64 blocks)")
	}
	if nblocks > blocksPerGroup {
		return layout{}, command.Failuref("filesystem too large for a single block group (max %d blocks)", blocksPerGroup)
	}
	inodes := nblocks * blockSize / inodeRatio
	if inodes < 16 {
		inodes = 16
	}
	inodes = roundUp(inodes, inodesPerBlock)
	inodeTableBlocks := inodes / inodesPerBlock

	blockBitmapBlock := firstDataBlock + 2 // after superblock(1) and GDT(2)
	inodeBitmapBlock := blockBitmapBlock + 1
	inodeTableBlock := inodeBitmapBlock + 1
	rootDirBlock := inodeTableBlock + inodeTableBlocks
	lostFoundBlock := rootDirBlock + 1

	return layout{
		blocks:           nblocks,
		inodes:           inodes,
		inodeTableBlocks: inodeTableBlocks,
		blockBitmapBlock: blockBitmapBlock,
		inodeBitmapBlock: inodeBitmapBlock,
		inodeTableBlock:  inodeTableBlock,
		rootDirBlock:     rootDirBlock,
		lostFoundBlock:   lostFoundBlock,
		usedBlocks:       lostFoundBlock, // blocks firstDataBlock..lostFoundBlock are used
		usedInodes:       firstInode,     // inodes 1..11 are used
	}, nil
}

// writeExt2 lays out the superblock, group descriptor, bitmaps, inode table,
// and the root and lost+found directories.
func writeExt2(f *os.File, nblocks int) error {
	l, err := computeLayout(nblocks)
	if err != nil {
		return err
	}
	groupBlocks := l.blocks - firstDataBlock // blocks firstDataBlock..blocks-1
	freeBlocks := groupBlocks - l.usedBlocks
	freeInodes := l.inodes - l.usedInodes
	reserved := l.blocks / 20
	ts := uint32(now().Unix())

	if err := writeAt(f, superblock(l, freeBlocks, freeInodes, reserved, ts), firstDataBlock*blockSize); err != nil {
		return err
	}
	if err := writeAt(f, groupDesc(l, freeBlocks, freeInodes), (firstDataBlock+1)*blockSize); err != nil {
		return err
	}
	if err := writeAt(f, blockBitmap(l, groupBlocks), int64(l.blockBitmapBlock)*blockSize); err != nil {
		return err
	}
	if err := writeAt(f, inodeBitmap(l), int64(l.inodeBitmapBlock)*blockSize); err != nil {
		return err
	}
	if err := writeAt(f, inodeTable(l, ts), int64(l.inodeTableBlock)*blockSize); err != nil {
		return err
	}
	if err := writeAt(f, rootDirData(l), int64(l.rootDirBlock)*blockSize); err != nil {
		return err
	}
	return writeAt(f, lostFoundData(), int64(l.lostFoundBlock)*blockSize)
}

func superblock(l layout, freeBlocks, freeInodes, reserved int, ts uint32) []byte {
	sb := make([]byte, blockSize)
	le32(sb, 0, uint32(l.inodes))
	le32(sb, 4, uint32(l.blocks))
	le32(sb, 8, uint32(reserved))
	le32(sb, 12, uint32(freeBlocks))
	le32(sb, 16, uint32(freeInodes))
	le32(sb, 20, firstDataBlock)
	le32(sb, 24, 0) // log_block_size: 1024 << 0
	le32(sb, 28, 0) // log_frag_size
	le32(sb, 32, blocksPerGroup)
	le32(sb, 36, blocksPerGroup) // frags_per_group
	le32(sb, 40, uint32(l.inodes))
	le32(sb, 44, ts)     // mtime
	le32(sb, 48, ts)     // wtime
	le16(sb, 52, 0)      // mnt_count
	le16(sb, 54, 0xFFFF) // max_mnt_count = -1
	le16(sb, 56, extMagic)
	le16(sb, 58, 1)  // state: clean
	le16(sb, 60, 1)  // errors: continue
	le16(sb, 62, 0)  // minor_rev
	le32(sb, 64, ts) // lastcheck
	le32(sb, 68, 0)  // checkinterval
	le32(sb, 72, 0)  // creator_os: Linux
	le32(sb, 76, 1)  // rev_level: dynamic
	le16(sb, 80, 0)  // def_resuid
	le16(sb, 82, 0)  // def_resgid
	le32(sb, 84, firstInode)
	le16(sb, 88, inodeSize)
	le16(sb, 90, 0) // block_group_nr
	le32(sb, 92, 0) // feature_compat
	le32(sb, 96, featIncompatFiletype)
	le32(sb, 100, 0) // feature_ro_compat
	return sb
}

func groupDesc(l layout, freeBlocks, freeInodes int) []byte {
	// The descriptor lives at the start of the GDT block.
	blk := make([]byte, blockSize)
	le32(blk, 0, uint32(l.blockBitmapBlock))
	le32(blk, 4, uint32(l.inodeBitmapBlock))
	le32(blk, 8, uint32(l.inodeTableBlock))
	le16(blk, 12, uint16(freeBlocks))
	le16(blk, 14, uint16(freeInodes))
	le16(blk, 16, 2) // used_dirs_count: root and lost+found
	return blk
}

func blockBitmap(l layout, groupBlocks int) []byte {
	bm := make([]byte, blockSize)
	// Bit i == block (firstDataBlock + i). Mark used metadata+data blocks.
	for b := 0; b < l.usedBlocks; b++ {
		setBit(bm, b)
	}
	// Mark blocks past the end of the group (non-existent) as used.
	for b := groupBlocks; b < blockSize*8; b++ {
		setBit(bm, b)
	}
	return bm
}

func inodeBitmap(l layout) []byte {
	bm := make([]byte, blockSize)
	for i := 0; i < l.usedInodes; i++ { // inodes 1..usedInodes -> bits 0..usedInodes-1
		setBit(bm, i)
	}
	for i := l.inodes; i < blockSize*8; i++ {
		setBit(bm, i)
	}
	return bm
}

func inodeTable(l layout, ts uint32) []byte {
	table := make([]byte, l.inodeTableBlocks*blockSize)
	writeInode(table, rootInode, dirInode(sIFDIR|0o755, 3, l.rootDirBlock, ts))
	writeInode(table, lostFoundInode, dirInode(sIFDIR|0o700, 2, l.lostFoundBlock, ts))
	return table
}

// dirInode builds a 128-byte directory inode occupying a single block.
func dirInode(mode uint16, links int, block int, ts uint32) []byte {
	ino := make([]byte, inodeSize)
	le16(ino, 0, mode)
	le32(ino, 4, blockSize)      // size
	le32(ino, 8, ts)             // atime
	le32(ino, 12, ts)            // ctime
	le32(ino, 16, ts)            // mtime
	le16(ino, 26, uint16(links)) // links_count
	le32(ino, 28, blockSize/512) // i_blocks in 512-byte units
	le32(ino, 40, uint32(block)) // i_block[0]
	return ino
}

// writeInode places a 128-byte inode at its 1-indexed slot.
func writeInode(table []byte, num int, ino []byte) {
	copy(table[(num-1)*inodeSize:], ino)
}

func rootDirData(l layout) []byte {
	blk := make([]byte, blockSize)
	off := putDirEntry(blk, 0, rootInode, 2, ".")
	off = putDirEntry(blk, off, rootInode, 2, "..")
	putDirEntryFill(blk, off, lostFoundInode, 2, "lost+found")
	return blk
}

func lostFoundData() []byte {
	blk := make([]byte, blockSize)
	off := putDirEntry(blk, 0, lostFoundInode, 2, ".")
	putDirEntryFill(blk, off, rootInode, 2, "..")
	return blk
}

// putDirEntry writes a directory entry with the minimal aligned rec_len and
// returns the next offset.
func putDirEntry(blk []byte, off int, inode uint32, fileType byte, name string) int {
	recLen := roundUp(8+len(name), 4)
	writeDirEntry(blk, off, inode, recLen, fileType, name)
	return off + recLen
}

// putDirEntryFill writes the last entry, extending rec_len to the block end.
func putDirEntryFill(blk []byte, off int, inode uint32, fileType byte, name string) {
	writeDirEntry(blk, off, inode, blockSize-off, fileType, name)
}

func writeDirEntry(blk []byte, off int, inode uint32, recLen int, fileType byte, name string) {
	le32(blk, off, inode)
	le16(blk, off+4, uint16(recLen))
	blk[off+6] = byte(len(name))
	blk[off+7] = fileType
	copy(blk[off+8:], name)
}

func writeAt(f *os.File, data []byte, off int64) error {
	_, err := f.WriteAt(data, off)
	return err
}

func le16(b []byte, off int, v uint16) { binary.LittleEndian.PutUint16(b[off:], v) }
func le32(b []byte, off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }

func setBit(bm []byte, n int) { bm[n/8] |= 1 << (uint(n) % 8) }

func parseUint(s string) (int, error) {
	if s == "" {
		return 0, command.Failuref("invalid block count: %q", s)
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, command.Failuref("invalid block count: %q", s)
		}
		n = n*10 + int(s[i]-'0')
	}
	return n, nil
}
