package kbd

import (
	"encoding/binary"
	"fmt"
	"io"
)

// PSF (PC Screen Font) magic numbers and header sizes for the two versions the
// Linux console understands.
var (
	psf1Magic = []byte{0x36, 0x04}
	psf2Magic = []byte{0x72, 0xb5, 0x4a, 0x86}
)

// PSF1 mode flags.
const (
	psf1Mode512    = 0x01 // 512 glyphs instead of 256
	psf1ModeHasTab = 0x02 // a unicode table follows
	psf1ModeHasSeq = 0x04
)

// Font is a decoded console font: its dimensions, glyph count and the raw glyph
// bitmap data. It abstracts over PSF1 and PSF2 so the loadfont/setfont applets
// can validate a font without writing it to a device.
type Font struct {
	// Version is 1 or 2.
	Version int
	// Width and Height are the glyph dimensions in pixels.
	Width, Height int
	// Length is the number of glyphs.
	Length int
	// CharSize is the number of bytes per glyph.
	CharSize int
	// HasUnicodeTable reports whether a unicode mapping table follows the
	// glyphs.
	HasUnicodeTable bool
	// Glyphs is the raw glyph bitmap data (Length*CharSize bytes).
	Glyphs []byte
}

// DecodeFont parses a PSF1 or PSF2 console font from r, validating the magic,
// header and that enough glyph data is present. It does not require or parse the
// trailing unicode table; it only records whether one is declared.
func DecodeFont(r io.Reader) (*Font, error) {
	head := make([]byte, 4)
	if _, err := io.ReadFull(r, head); err != nil {
		return nil, fmt.Errorf("reading font header: %w", err)
	}
	switch {
	case head[0] == psf1Magic[0] && head[1] == psf1Magic[1]:
		return decodePSF1(head, r)
	case head[0] == psf2Magic[0] && head[1] == psf2Magic[1] &&
		head[2] == psf2Magic[2] && head[3] == psf2Magic[3]:
		return decodePSF2(r)
	default:
		return nil, fmt.Errorf("not a PSF console font: bad magic % x", head)
	}
}

// decodePSF1 parses a PSF1 font. head holds the first 4 bytes already read; for
// PSF1 those are magic[0], magic[1], mode, charsize.
func decodePSF1(head []byte, r io.Reader) (*Font, error) {
	mode := head[2]
	charSize := int(head[3])
	if charSize == 0 {
		return nil, fmt.Errorf("invalid PSF1 charsize 0")
	}
	length := 256
	if mode&psf1Mode512 != 0 {
		length = 512
	}
	f := &Font{
		Version:         1,
		Width:           8, // PSF1 glyphs are always 8 pixels wide
		Height:          charSize,
		Length:          length,
		CharSize:        charSize,
		HasUnicodeTable: mode&(psf1ModeHasTab|psf1ModeHasSeq) != 0,
	}
	return readGlyphs(f, r)
}

// psf2Header mirrors the fixed PSF2 header following the 4-byte magic.
type psf2Header struct {
	Version    uint32
	HeaderSize uint32
	Flags      uint32
	Length     uint32
	CharSize   uint32
	Height     uint32
	Width      uint32
}

const psf2HasUnicodeTable = 0x01

// decodePSF2 parses a PSF2 font (the 4-byte magic has already been consumed).
func decodePSF2(r io.Reader) (*Font, error) {
	var h psf2Header
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("reading PSF2 header: %w", err)
	}
	if h.CharSize == 0 || h.Length == 0 {
		return nil, fmt.Errorf("invalid PSF2 header: length=%d charsize=%d", h.Length, h.CharSize)
	}
	// Skip any header padding past the fixed 32-byte (magic+header) prefix.
	const fixed = 32
	if h.HeaderSize > fixed {
		if _, err := io.CopyN(io.Discard, r, int64(h.HeaderSize)-fixed); err != nil {
			return nil, fmt.Errorf("skipping PSF2 header padding: %w", err)
		}
	}
	f := &Font{
		Version:         2,
		Width:           int(h.Width),
		Height:          int(h.Height),
		Length:          int(h.Length),
		CharSize:        int(h.CharSize),
		HasUnicodeTable: h.Flags&psf2HasUnicodeTable != 0,
	}
	return readGlyphs(f, r)
}

// readGlyphs reads Length*CharSize bytes of glyph data into f and returns it.
func readGlyphs(f *Font, r io.Reader) (*Font, error) {
	n := f.Length * f.CharSize
	f.Glyphs = make([]byte, n)
	if _, err := io.ReadFull(r, f.Glyphs); err != nil {
		return nil, fmt.Errorf("reading %d bytes of glyph data: %w", n, err)
	}
	return f, nil
}

// Describe returns a one-line human summary of the font, used by --help-style
// dry-run output and by tests.
func (f *Font) Describe() string {
	tab := "no"
	if f.HasUnicodeTable {
		tab = "yes"
	}
	return fmt.Sprintf("PSF%d font: %d glyphs, %dx%d px, %d bytes/glyph, unicode table: %s",
		f.Version, f.Length, f.Width, f.Height, f.CharSize, tab)
}
