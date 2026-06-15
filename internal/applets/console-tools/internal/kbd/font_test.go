package kbd

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func buildPSF1(t *testing.T, mode byte, charSize int) []byte {
	t.Helper()
	var b bytes.Buffer
	b.Write(psf1Magic)
	b.WriteByte(mode)
	b.WriteByte(byte(charSize))
	length := 256
	if mode&psf1Mode512 != 0 {
		length = 512
	}
	b.Write(make([]byte, length*charSize))
	return b.Bytes()
}

func buildPSF2(t *testing.T, length, width, height, charSize int, flags uint32) []byte {
	t.Helper()
	var b bytes.Buffer
	b.Write(psf2Magic)
	h := psf2Header{
		Version:    0,
		HeaderSize: 32,
		Flags:      flags,
		Length:     uint32(length),
		CharSize:   uint32(charSize),
		Height:     uint32(height),
		Width:      uint32(width),
	}
	_ = binary.Write(&b, binary.LittleEndian, &h)
	b.Write(make([]byte, length*charSize))
	return b.Bytes()
}

func TestDecodePSF1(t *testing.T) {
	t.Parallel()
	f, err := DecodeFont(bytes.NewReader(buildPSF1(t, psf1ModeHasTab, 16)))
	if err != nil {
		t.Fatalf("DecodeFont: %v", err)
	}
	if f.Version != 1 || f.Length != 256 || f.Height != 16 || f.Width != 8 || f.CharSize != 16 {
		t.Errorf("unexpected font: %+v", f)
	}
	if !f.HasUnicodeTable {
		t.Error("expected unicode table flag")
	}
}

func TestDecodePSF1_512(t *testing.T) {
	t.Parallel()
	f, err := DecodeFont(bytes.NewReader(buildPSF1(t, psf1Mode512, 8)))
	if err != nil {
		t.Fatalf("DecodeFont: %v", err)
	}
	if f.Length != 512 {
		t.Errorf("Length = %d, want 512", f.Length)
	}
}

func TestDecodePSF2(t *testing.T) {
	t.Parallel()
	f, err := DecodeFont(bytes.NewReader(buildPSF2(t, 256, 8, 16, 16, psf2HasUnicodeTable)))
	if err != nil {
		t.Fatalf("DecodeFont: %v", err)
	}
	if f.Version != 2 || f.Length != 256 || f.Width != 8 || f.Height != 16 {
		t.Errorf("unexpected font: %+v", f)
	}
	if !f.HasUnicodeTable {
		t.Error("expected unicode table flag")
	}
	if !strings.Contains(f.Describe(), "PSF2 font: 256 glyphs") {
		t.Errorf("Describe = %q", f.Describe())
	}
}

func TestDecodeFontBadMagic(t *testing.T) {
	t.Parallel()
	if _, err := DecodeFont(strings.NewReader("XXXXdata")); err == nil {
		t.Fatal("expected bad magic error")
	}
}

func TestDecodeFontTruncatedGlyphs(t *testing.T) {
	t.Parallel()
	data := buildPSF2(t, 256, 8, 16, 16, 0)
	if _, err := DecodeFont(bytes.NewReader(data[:40])); err == nil {
		t.Fatal("expected truncated glyph error")
	}
}

func TestDecodePSF1ZeroCharsize(t *testing.T) {
	t.Parallel()
	data := []byte{psf1Magic[0], psf1Magic[1], 0, 0}
	if _, err := DecodeFont(bytes.NewReader(data)); err == nil {
		t.Fatal("expected error for charsize 0")
	}
}
