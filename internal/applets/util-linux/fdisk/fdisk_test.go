package fdisk

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

// craftMBR builds a 512-byte MBR image; each part is {bootable, type, start, sectors}.
func craftMBR(t *testing.T, withSig bool, parts ...[4]uint32) string {
	t.Helper()
	mbr := make([]byte, sectorSize)
	for i, p := range parts {
		e := tableOffset + i*entrySize
		if p[0] != 0 {
			mbr[e] = bootableFlag
		}
		mbr[e+4] = byte(p[1])
		binary.LittleEndian.PutUint32(mbr[e+8:], p[2])
		binary.LittleEndian.PutUint32(mbr[e+12:], p[3])
	}
	if withSig {
		mbr[signatureLow] = 0x55
		mbr[signatureLow+1] = 0xAA
	}
	p := filepath.Join(t.TempDir(), "disk.img")
	if err := os.WriteFile(p, mbr, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestListsPartitions(t *testing.T) {
	img := craftMBR(t, true,
		[4]uint32{1, 0x83, 2048, 4192256},  // bootable Linux
		[4]uint32{0, 0x82, 4194304, 2097152}, // swap
	)
	out, err := run(t, "-l", img)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2048") || !strings.Contains(out, "4194303") || !strings.Contains(out, "Linux") {
		t.Errorf("first partition wrong:\n%s", out)
	}
	if !strings.Contains(out, "Linux swap") {
		t.Errorf("swap partition wrong:\n%s", out)
	}
	// The bootable flag column must show '*' for the first partition only.
	if strings.Count(out, "*") != 1 {
		t.Errorf("boot flag count wrong:\n%s", out)
	}
}

func TestUnknownTypeAndEmptySkipped(t *testing.T) {
	img := craftMBR(t, true,
		[4]uint32{0, 0xAB, 100, 200}, // unknown type
		[4]uint32{0, 0x00, 0, 0},     // empty -> skipped
	)
	out, err := run(t, "-l", img)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "unknown (0xab)") {
		t.Errorf("unknown type not labeled:\n%s", out)
	}
	// Only one data row (the empty entry is skipped): header + one row.
	if lines := strings.Count(strings.TrimSpace(out), "\n"); lines != 1 {
		t.Errorf("expected one partition row, got:\n%s", out)
	}
}

func TestBadSignature(t *testing.T) {
	img := craftMBR(t, false, [4]uint32{0, 0x83, 2048, 1000})
	if _, err := run(t, "-l", img); err == nil {
		t.Errorf("a missing MBR signature should fail")
	}
}

func TestRequiresListMode(t *testing.T) {
	img := craftMBR(t, true, [4]uint32{0, 0x83, 2048, 1000})
	if _, err := run(t, img); err == nil {
		t.Errorf("without -l it should fail")
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t, "-l"); err == nil {
		t.Errorf("missing device should fail")
	}
	if _, err := run(t, "-l", "/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
}
