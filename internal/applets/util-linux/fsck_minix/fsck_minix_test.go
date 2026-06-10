package fsckminix

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

func craft(t *testing.T, magic uint16, ninodes, nzones, firstZone uint16) string {
	t.Helper()
	buf := make([]byte, superblockOffset+blockSize)
	sb := buf[superblockOffset:]
	binary.LittleEndian.PutUint16(sb[0:], ninodes)
	binary.LittleEndian.PutUint16(sb[2:], nzones)
	binary.LittleEndian.PutUint16(sb[8:], firstZone)
	binary.LittleEndian.PutUint16(sb[16:], magic)
	p := filepath.Join(t.TempDir(), "fs.img")
	if err := os.WriteFile(p, buf, 0o644); err != nil {
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

func TestValidV1(t *testing.T) {
	img := craft(t, 0x137F, 704, 2048, 26)
	out, err := run(t, img)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Minix v1, 14-character names", "704 inodes", "2048 zones", "firstdatazone=26"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestValidV2(t *testing.T) {
	img := craft(t, 0x2478, 100, 500, 10)
	out, err := run(t, img)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Minix v2, 30-character names") {
		t.Errorf("v2 not recognized:\n%s", out)
	}
}

func TestBadMagic(t *testing.T) {
	img := craft(t, 0x0000, 0, 0, 0)
	if _, err := run(t, img); err == nil {
		t.Errorf("a bad magic should fail")
	}
}

func TestForceFlagAccepted(t *testing.T) {
	img := craft(t, 0x137F, 1, 16, 3)
	if _, err := run(t, "-f", img); err != nil {
		t.Errorf("-f should be accepted: %v", err)
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("missing device should fail")
	}
	if _, err := run(t, "/no/such/image"); err == nil {
		t.Errorf("a missing image should fail")
	}
}
