package modutils

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

func run(t *testing.T, c *Command, args ...string) (string, string, error) {
	t.Helper()
	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := c.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeKO(t *testing.T, dir, name string, modinfo []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, buildELF(modinfo), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestValidateModuleName(t *testing.T) {
	t.Parallel()
	if err := validateModuleName("loop"); err != nil {
		t.Errorf("loop should be valid: %v", err)
	}
	for _, bad := range []string{"", "loop.ko", "/lib/loop", "dir/loop"} {
		if err := validateModuleName(bad); err == nil {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

func TestInsmodValidatesThenGates(t *testing.T) {
	dir := t.TempDir()
	ko := writeKO(t, dir, "loop.ko", []byte("license=GPL\x00depends=\x00"))
	_, _, err := run(t, NewInsmod(), ko)
	if err == nil {
		t.Fatal("expected gated error")
	}
	if !strings.Contains(err.Error(), "CAP_SYS_MODULE") || !strings.Contains(err.Error(), "validated successfully") {
		t.Errorf("error lacks documented gating: %v", err)
	}
}

func TestInsmodRejectsNonModule(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.ko")
	if err := os.WriteFile(bad, []byte("nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, NewInsmod(), bad); err == nil {
		t.Fatal("expected error for non-ELF module")
	}
}

func TestRmmodGates(t *testing.T) {
	_, _, err := run(t, NewRmmod(), "loop")
	if err == nil || !strings.Contains(err.Error(), "CAP_SYS_MODULE") {
		t.Errorf("rmmod should gate: %v", err)
	}
	if _, _, err := run(t, NewRmmod(), "loop.ko"); err == nil {
		t.Error("rmmod should reject path-like names")
	}
}

func TestModprobeGates(t *testing.T) {
	_, _, err := run(t, NewModprobe(), "loop")
	if err == nil || !strings.Contains(err.Error(), "CAP_SYS_MODULE") {
		t.Errorf("modprobe should gate: %v", err)
	}
}

func TestDepmodDryRun(t *testing.T) {
	dir := t.TempDir()
	writeKO(t, dir, "ext4.ko", []byte("depends=mbcache,jbd2\x00"))
	writeKO(t, dir, "loop.ko", []byte("depends=\x00"))

	out, _, err := run(t, NewDepmod(), "-n", dir)
	if err != nil {
		t.Fatalf("depmod -n: %v", err)
	}
	if !strings.Contains(out, "ext4.ko: mbcache jbd2") {
		t.Errorf("ext4 deps missing:\n%s", out)
	}
	if !strings.Contains(out, "loop.ko:") {
		t.Errorf("loop entry missing:\n%s", out)
	}
}

func TestDepmodInstallGated(t *testing.T) {
	dir := t.TempDir()
	writeKO(t, dir, "loop.ko", []byte("depends=\x00"))
	_, _, err := run(t, NewDepmod(), dir)
	if err == nil || !strings.Contains(err.Error(), "modules.dep") {
		t.Errorf("depmod install should gate: %v", err)
	}
}

// buildELF returns a minimal 64-bit little-endian ELF with a single ".modinfo"
// section, enough for debug/elf to parse it.
func buildELF(modinfo []byte) []byte {
	const ehSize = 64
	shstrtab := []byte("\x00.modinfo\x00.shstrtab\x00")
	modinfoNameOff := uint32(1)
	shstrtabNameOff := uint32(1 + len(".modinfo") + 1)

	modinfoOff := uint64(ehSize)
	shstrtabOff := modinfoOff + uint64(len(modinfo))
	shoff := shstrtabOff + uint64(len(shstrtab))

	buf := &bytes.Buffer{}
	buf.Write([]byte{0x7f, 'E', 'L', 'F'})
	buf.WriteByte(2)
	buf.WriteByte(1)
	buf.WriteByte(1)
	buf.WriteByte(0)
	buf.Write(make([]byte, 8))
	le := binary.LittleEndian
	w16 := func(v uint16) { _ = binary.Write(buf, le, v) }
	w32 := func(v uint32) { _ = binary.Write(buf, le, v) }
	w64 := func(v uint64) { _ = binary.Write(buf, le, v) }
	w16(1)
	w16(62)
	w32(1)
	w64(0)
	w64(0)
	w64(shoff)
	w32(0)
	w16(ehSize)
	w16(0)
	w16(0)
	w16(64)
	w16(3)
	w16(2)

	buf.Write(modinfo)
	buf.Write(shstrtab)

	writeSH := func(name uint32, typ uint32, off uint64, size uint64) {
		w32(name)
		w32(typ)
		w64(0)
		w64(0)
		w64(off)
		w64(size)
		w32(0)
		w32(0)
		w64(1)
		w64(0)
	}
	writeSH(0, 0, 0, 0)
	writeSH(modinfoNameOff, 1, modinfoOff, uint64(len(modinfo)))
	writeSH(shstrtabNameOff, 3, shstrtabOff, uint64(len(shstrtab)))
	return buf.Bytes()
}
