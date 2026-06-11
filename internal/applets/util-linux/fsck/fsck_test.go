package fsck

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

func TestDetect(t *testing.T) {
	t.Parallel()

	ext := make([]byte, 2048)
	binary.LittleEndian.PutUint16(ext[1024+56:], 0xEF53)
	if got := detect(ext); got != "ext2/ext3/ext4" {
		t.Errorf("ext = %q", got)
	}

	minix := make([]byte, 2048)
	binary.LittleEndian.PutUint16(minix[1024+16:], 0x137F)
	if got := detect(minix); got != "minix" {
		t.Errorf("minix = %q", got)
	}

	vfat := make([]byte, 512)
	vfat[510], vfat[511] = 0x55, 0xAA
	copy(vfat[54:], "FAT16   ")
	if got := detect(vfat); got != "vfat" {
		t.Errorf("vfat = %q", got)
	}

	// FAT32 keeps its type label at offset 82.
	vfat32 := make([]byte, 512)
	vfat32[510], vfat32[511] = 0x55, 0xAA
	copy(vfat32[82:], "FAT32   ")
	if got := detect(vfat32); got != "vfat" {
		t.Errorf("vfat32 = %q", got)
	}

	swap := make([]byte, 4096)
	copy(swap[4086:], "SWAPSPACE2")
	if got := detect(swap); got != "swap" {
		t.Errorf("swap = %q", got)
	}

	if got := detect(make([]byte, 8192)); got != "" {
		t.Errorf("zeros should be unrecognized, got %q", got)
	}
}

func TestRunReportsType(t *testing.T) {
	buf := make([]byte, 2048)
	binary.LittleEndian.PutUint16(buf[1024+16:], 0x2478) // minix v2, 30-char
	img := filepath.Join(t.TempDir(), "fs.img")
	if err := os.WriteFile(img, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{img}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "minix") {
		t.Errorf("report = %q", out.String())
	}
}

func TestRunErrors(t *testing.T) {
	run := func(args ...string) error {
		io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		return New().Run(context.Background(), io, args)
	}
	if err := run(); err == nil {
		t.Errorf("missing device should fail")
	}
	if err := run("/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
	// An unrecognized image is an error.
	img := filepath.Join(t.TempDir(), "u.img")
	if err := os.WriteFile(img, make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(img); err == nil {
		t.Errorf("an unrecognized filesystem should fail")
	}
}
