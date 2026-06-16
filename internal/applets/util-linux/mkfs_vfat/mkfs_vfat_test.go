package mkfsvfat

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

func makeImage(t *testing.T, kib int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "fat.img")
	if err := os.WriteFile(p, make([]byte, kib*1024), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestBootSector(t *testing.T) {
	img := makeImage(t, 8192)
	if err := run(t, img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	if binary.LittleEndian.Uint16(data[11:]) != sectorSize {
		t.Errorf("bytes per sector wrong")
	}
	if data[13] != 1 || data[16] != numFATs {
		t.Errorf("spc=%d nfats=%d", data[13], data[16])
	}
	if binary.LittleEndian.Uint16(data[17:]) != rootEntries {
		t.Errorf("root entries wrong")
	}
	if data[21] != mediaByte {
		t.Errorf("media byte = %#x", data[21])
	}
	if string(data[54:62]) != "FAT16   " {
		t.Errorf("fs type = %q", data[54:62])
	}
	if data[510] != 0x55 || data[511] != 0xAA {
		t.Errorf("boot signature missing")
	}
	// FAT starts with the media byte and end-of-chain markers.
	fatOff := reservedSectors * sectorSize
	if data[fatOff] != mediaByte || data[fatOff+1] != 0xFF {
		t.Errorf("FAT head wrong: % x", data[fatOff:fatOff+4])
	}
}

func TestVolumeLabelEntry(t *testing.T) {
	img := makeImage(t, 8192)
	if err := run(t, "-n", "MYVOL", img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	if string(data[43:48]) != "MYVOL" {
		t.Errorf("boot label = %q", data[43:54])
	}
	// Root directory volume-label entry.
	g, _ := computeGeometry(16384)
	rootOff := (reservedSectors + numFATs*g.sectorsPerFAT) * sectorSize
	if string(data[rootOff:rootOff+5]) != "MYVOL" || data[rootOff+11] != 0x08 {
		t.Errorf("root volume entry wrong")
	}
}

func TestNoLabelNoRootEntry(t *testing.T) {
	img := makeImage(t, 8192)
	if err := run(t, img); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(img)
	g, _ := computeGeometry(16384)
	rootOff := (reservedSectors + numFATs*g.sectorsPerFAT) * sectorSize
	if data[rootOff] != 0 {
		t.Errorf("root dir should be empty without -n")
	}
}

func TestTooSmall(t *testing.T) {
	img := makeImage(t, 1024) // below the FAT16 cluster minimum
	if err := run(t, img); err == nil {
		t.Errorf("a too-small image should be rejected")
	}
}

func TestMkdosfsAlias(t *testing.T) {
	if New().Name() != "mkfs.vfat" || NewMkdosfs().Name() != "mkdosfs" {
		t.Errorf("alias names wrong: %q / %q", New().Name(), NewMkdosfs().Name())
	}
	// Both aliases resolve their synopsis through the shared name -> config table.
	for _, c := range []*Command{New(), NewMkdosfs()} {
		if c.Synopsis() != "Create a FAT16 filesystem" {
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

func TestComputeGeometryClusterRange(t *testing.T) {
	t.Parallel()
	// Small device: spc 1, clusters in the FAT16 range.
	g, err := computeGeometry(16384)
	if err != nil || g.sectorsPerCluster != 1 {
		t.Fatalf("small geometry = %+v, err %v", g, err)
	}
	if g.clusters < minFAT16Clusters || g.clusters > maxFAT16Clusters {
		t.Errorf("clusters %d out of FAT16 range", g.clusters)
	}
	// Large device (200 MiB): a bigger cluster keeps the count within FAT16.
	big, err := computeGeometry(409600)
	if err != nil {
		t.Fatalf("large geometry err: %v", err)
	}
	if big.sectorsPerCluster <= 1 {
		t.Errorf("expected spc>1 for a large device, got %d", big.sectorsPerCluster)
	}
	if big.clusters > maxFAT16Clusters {
		t.Errorf("large clusters %d exceed FAT16 max", big.clusters)
	}
	// Way too small: an error rather than silent FAT12.
	if _, err := computeGeometry(64); err == nil {
		t.Errorf("a tiny device should error")
	}
}
