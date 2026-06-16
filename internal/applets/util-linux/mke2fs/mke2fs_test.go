package mke2fs

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func makeImage(t *testing.T, blocks int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "ext2.img")
	if err := os.WriteFile(p, make([]byte, blocks*blockSize), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	on := now
	now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	defer func() { now = on }()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestSuperblock(t *testing.T) {
	img := makeImage(t, 1024)
	if err := run(t, img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	sb := data[firstDataBlock*blockSize:]

	if binary.LittleEndian.Uint16(sb[56:]) != extMagic {
		t.Errorf("magic wrong")
	}
	if binary.LittleEndian.Uint32(sb[4:]) != 1024 {
		t.Errorf("block count = %d", binary.LittleEndian.Uint32(sb[4:]))
	}
	if got := binary.LittleEndian.Uint16(sb[58:]); got != 1 {
		t.Errorf("state = %d, want 1 (clean)", got)
	}
	if got := binary.LittleEndian.Uint32(sb[84:]); got != firstInode {
		t.Errorf("first inode = %d", got)
	}
	if got := binary.LittleEndian.Uint16(sb[88:]); got != inodeSize {
		t.Errorf("inode size = %d", got)
	}
	if got := binary.LittleEndian.Uint32(sb[76:]); got != 1 {
		t.Errorf("rev level = %d, want 1", got)
	}
	// free counts must be internally consistent.
	l, _ := computeLayout(1024)
	wantFreeBlocks := (1024 - firstDataBlock) - l.usedBlocks
	if got := int(binary.LittleEndian.Uint32(sb[12:])); got != wantFreeBlocks {
		t.Errorf("free blocks = %d, want %d", got, wantFreeBlocks)
	}
	if got := int(binary.LittleEndian.Uint32(sb[16:])); got != l.inodes-firstInode {
		t.Errorf("free inodes = %d, want %d", got, l.inodes-firstInode)
	}
}

func TestRootDirectory(t *testing.T) {
	img := makeImage(t, 1024)
	if err := run(t, img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	l, _ := computeLayout(1024)

	// Root inode (#2) must be a directory.
	inodeOff := l.inodeTableBlock*blockSize + (rootInode-1)*inodeSize
	mode := binary.LittleEndian.Uint16(data[inodeOff:])
	if mode&sIFDIR == 0 {
		t.Errorf("root inode mode %#o not a directory", mode)
	}

	// Root directory entries: ".", "..", "lost+found".
	dir := data[l.rootDirBlock*blockSize:]
	names := dirNames(dir)
	for _, want := range []string{".", "..", "lost+found"} {
		if !names[want] {
			t.Errorf("root dir missing %q (have %v)", want, names)
		}
	}
}

func dirNames(dir []byte) map[string]bool {
	names := map[string]bool{}
	off := 0
	for off+8 <= len(dir) {
		recLen := int(binary.LittleEndian.Uint16(dir[off+4:]))
		if recLen < 8 {
			break
		}
		nameLen := int(dir[off+6])
		if inode := binary.LittleEndian.Uint32(dir[off:]); inode != 0 && nameLen > 0 {
			names[string(dir[off+8:off+8+nameLen])] = true
		}
		off += recLen
	}
	return names
}

func TestBounds(t *testing.T) {
	if err := run(t, makeImage(t, 32)); err == nil {
		t.Errorf("a too-small image should fail")
	}
	if err := run(t, makeImage(t, 10000)); err == nil {
		t.Errorf("a too-large image should fail")
	}
}

func TestBlocksOperand(t *testing.T) {
	img := makeImage(t, 4096)
	if err := run(t, img, "1024"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	if got := binary.LittleEndian.Uint32(data[firstDataBlock*blockSize+4:]); got != 1024 {
		t.Errorf("block count from operand = %d, want 1024", got)
	}
}

func TestAlias(t *testing.T) {
	if New().Name() != "mke2fs" || NewMkfsExt2().Name() != "mkfs.ext2" {
		t.Errorf("alias names wrong")
	}
	// Both aliases resolve their synopsis through the shared name -> config table.
	for _, c := range []*Command{New(), NewMkfsExt2()} {
		if c.Synopsis() != "Create an ext2 filesystem" {
			t.Errorf("%s synopsis = %q", c.Name(), c.Synopsis())
		}
	}
}

func TestErrors(t *testing.T) {
	if err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
	if err := run(t, "/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
}
