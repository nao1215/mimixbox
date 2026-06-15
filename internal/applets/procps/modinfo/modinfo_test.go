package modinfo

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseModinfo(t *testing.T) {
	t.Parallel()
	data := []byte("license=GPL\x00author=Linus\x00depends=\x00alias=block-major-7-*\x00garbagewithoutequals\x00")
	pairs := parseModinfo(data)
	want := []pair{
		{"license", "GPL"},
		{"author", "Linus"},
		{"depends", ""},
		{"alias", "block-major-7-*"},
	}
	if !reflect.DeepEqual(pairs, want) {
		t.Fatalf("parseModinfo = %+v, want %+v", pairs, want)
	}
	if got := fieldNames(pairs); !reflect.DeepEqual(got, []string{"alias", "author", "depends", "license"}) {
		t.Errorf("fieldNames = %v", got)
	}
}

func TestRunWithFixtureELF(t *testing.T) {
	dir := t.TempDir()
	ko := filepath.Join(dir, "loop.ko")
	modinfo := []byte("license=GPL\x00version=1.0\x00description=Loopback device\x00")
	if err := os.WriteFile(ko, buildELF(modinfo), 0o600); err != nil {
		t.Fatal(err)
	}

	// All fields.
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{ko}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := out.String()
	for _, want := range []string{"filename:", ko, "license:", "GPL", "version:", "description:", "Loopback device"} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q:\n%s", want, s)
		}
	}

	// Single field with -F.
	out.Reset()
	if err := New().Run(context.Background(), io, []string{"-F", "license", ko}); err != nil {
		t.Fatalf("Run -F: %v", err)
	}
	if strings.TrimSpace(out.String()) != "GPL" {
		t.Errorf("-F license = %q, want GPL", out.String())
	}
}

func TestRunNonELF(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.ko")
	if err := os.WriteFile(bad, []byte("not an elf"), 0o600); err != nil {
		t.Fatal(err)
	}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: errBuf}
	if err := New().Run(context.Background(), io, []string{bad}); err == nil {
		t.Fatal("expected error for non-ELF file")
	}
}

func TestRunNoArgs(t *testing.T) {
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error with no files")
	}
}

// buildELF returns a minimal but valid 64-bit little-endian ELF object that
// contains a single ".modinfo" section holding modinfo. It is just enough for
// debug/elf to parse and expose the section.
func buildELF(modinfo []byte) []byte {
	const (
		ehSize = 64
		shSize = 64
	)
	// Layout: [ELF header][.modinfo data][.shstrtab data][section headers]
	shstrtab := []byte("\x00.modinfo\x00.shstrtab\x00")
	modinfoNameOff := uint32(1) // ".modinfo"
	shstrtabNameOff := uint32(1 + len(".modinfo") + 1)

	modinfoOff := uint64(ehSize)
	shstrtabOff := modinfoOff + uint64(len(modinfo))
	shoff := shstrtabOff + uint64(len(shstrtab))

	buf := &bytes.Buffer{}
	// e_ident
	buf.Write([]byte{0x7f, 'E', 'L', 'F'})
	buf.WriteByte(2) // ELFCLASS64
	buf.WriteByte(1) // ELFDATA2LSB
	buf.WriteByte(1) // EV_CURRENT
	buf.WriteByte(0) // OSABI
	buf.Write(make([]byte, 8))
	le := binary.LittleEndian
	w16 := func(v uint16) { _ = binary.Write(buf, le, v) }
	w32 := func(v uint32) { _ = binary.Write(buf, le, v) }
	w64 := func(v uint64) { _ = binary.Write(buf, le, v) }
	w16(1)      // e_type = ET_REL
	w16(62)     // e_machine = x86-64
	w32(1)      // e_version
	w64(0)      // e_entry
	w64(0)      // e_phoff
	w64(shoff)  // e_shoff
	w32(0)      // e_flags
	w16(ehSize) // e_ehsize
	w16(0)      // e_phentsize
	w16(0)      // e_phnum
	w16(shSize) // e_shentsize
	w16(3)      // e_shnum (null, .modinfo, .shstrtab)
	w16(2)      // e_shstrndx

	buf.Write(modinfo)
	buf.Write(shstrtab)

	writeSH := func(name uint32, typ uint32, off uint64, size uint64) {
		w32(name)
		w32(typ)
		w64(0)   // flags
		w64(0)   // addr
		w64(off) // offset
		w64(size)
		w32(0) // link
		w32(0) // info
		w64(1) // addralign
		w64(0) // entsize
	}
	// Section 0: null
	writeSH(0, 0, 0, 0)
	// Section 1: .modinfo (PROGBITS=1)
	writeSH(modinfoNameOff, 1, modinfoOff, uint64(len(modinfo)))
	// Section 2: .shstrtab (STRTAB=3)
	writeSH(shstrtabNameOff, 3, shstrtabOff, uint64(len(shstrtab)))

	return buf.Bytes()
}
