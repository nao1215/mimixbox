package mkfsminix

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func makeImage(t *testing.T, blocks int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "minix.img")
	if err := os.WriteFile(p, make([]byte, blocks*blockSize), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestCreatesValidSuperblock(t *testing.T) {
	img := makeImage(t, 2048)
	if err := run(t, img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	sb := data[blockSize:] // superblock is block 1

	if magic := binary.LittleEndian.Uint16(sb[16:]); magic != minixMagic {
		t.Errorf("magic = %#x, want %#x", magic, minixMagic)
	}
	if state := binary.LittleEndian.Uint16(sb[18:]); state != 1 {
		t.Errorf("state = %d, want 1 (clean)", state)
	}
	ninodes := binary.LittleEndian.Uint16(sb[0:])
	nzones := binary.LittleEndian.Uint16(sb[2:])
	if ninodes == 0 || nzones != 2048 {
		t.Errorf("ninodes=%d nzones=%d", ninodes, nzones)
	}
	firstZone := int(binary.LittleEndian.Uint16(sb[8:]))
	if firstZone < 3 {
		t.Errorf("firstdatazone = %d, too small", firstZone)
	}

	// Root inode (inode 1) at the start of the inode table.
	l := computeLayout(2048)
	inodeOff := (2 + l.imapBlocks + l.zmapBlocks) * blockSize
	mode := binary.LittleEndian.Uint16(data[inodeOff:])
	if mode&sIFDIR == 0 {
		t.Errorf("root inode mode %#o is not a directory", mode)
	}

	// Root directory must contain "." and ".." pointing at inode 1.
	dirOff := firstZone * blockSize
	if binary.LittleEndian.Uint16(data[dirOff:]) != 1 || string(bytes.TrimRight(data[dirOff+2:dirOff+16], "\x00")) != "." {
		t.Errorf("first dir entry wrong")
	}
	if binary.LittleEndian.Uint16(data[dirOff+16:]) != 1 || string(bytes.TrimRight(data[dirOff+18:dirOff+32], "\x00")) != ".." {
		t.Errorf("second dir entry wrong")
	}
}

func TestBlocksOperand(t *testing.T) {
	// A larger file but an explicit smaller block count must be honored.
	img := makeImage(t, 4096)
	if err := run(t, img, "1024"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	if nzones := binary.LittleEndian.Uint16(data[blockSize+2:]); nzones != 1024 {
		t.Errorf("nzones = %d, want 1024 (from operand)", nzones)
	}
}

func TestLayoutConsistency(t *testing.T) {
	t.Parallel()
	l := computeLayout(8192)
	// firstDataZone must sit after the boot block, superblock, bitmaps, inodes.
	want := 2 + l.imapBlocks + l.zmapBlocks + l.inodeBlocks
	if l.firstDataZone != want {
		t.Errorf("firstDataZone = %d, want %d", l.firstDataZone, want)
	}
	if l.firstDataZone >= l.nzones {
		t.Errorf("firstDataZone %d >= nzones %d", l.firstDataZone, l.nzones)
	}
}

func TestErrors(t *testing.T) {
	if err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
	if err := run(t, "/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
	small := makeImage(t, 4) // below minBlocks
	if err := run(t, small); err == nil {
		t.Errorf("a too-small device should fail")
	}
}
