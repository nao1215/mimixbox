// Package rpmfile parses just enough of the RPM package format to drive the
// rpm2cpio and rpm applets: the 96-byte lead, the signature and main headers
// (tag/type/offset index plus the data store) and the compressed cpio payload.
package rpmfile

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Common RPM header tags (from the main header).
const (
	TagName       = 1000
	TagVersion    = 1001
	TagRelease    = 1002
	TagSummary    = 1004
	TagArch       = 1022
	TagDirindexes = 1116
	TagBasenames  = 1117
	TagDirnames   = 1118
)

// RPM index entry data types.
const (
	typeInt16       = 3
	typeInt32       = 4
	typeString      = 6
	typeStringArray = 8
	typeI18NString  = 9
)

const leadSize = 96

var headerMagic = []byte{0x8e, 0xad, 0xe8}

// entry is one parsed index entry.
type entry struct {
	tag    int32
	typ    int32
	offset int32
	count  int32
}

// Header is a parsed RPM header: its index plus the raw data store.
type Header struct {
	entries map[int32]entry
	store   []byte
}

// File is a parsed RPM file: the main header and a reader positioned at the
// (still compressed) payload.
type File struct {
	Header  *Header
	payload io.Reader
}

// Open parses r as an RPM file: it skips the lead and signature header, parses
// the main header and leaves the payload ready to read.
func Open(r io.Reader) (*File, error) {
	br := bufio.NewReader(r)

	lead := make([]byte, leadSize)
	if _, err := io.ReadFull(br, lead); err != nil {
		return nil, fmt.Errorf("short lead: %w", err)
	}
	if lead[0] != 0xed || lead[1] != 0xab || lead[2] != 0xee || lead[3] != 0xdb {
		return nil, fmt.Errorf("not an RPM file")
	}

	// Signature header is padded to an 8-byte boundary.
	if _, _, err := readHeader(br, true); err != nil {
		return nil, fmt.Errorf("signature header: %w", err)
	}
	h, _, err := readHeader(br, false)
	if err != nil {
		return nil, fmt.Errorf("main header: %w", err)
	}

	return &File{Header: h, payload: br}, nil
}

// Payload returns a reader over the decompressed cpio payload, detecting the
// compression from its magic bytes.
func (f *File) Payload() (io.Reader, error) {
	br := bufio.NewReader(f.payload)
	magic, err := br.Peek(6)
	if err != nil && err != io.EOF {
		return nil, err
	}
	switch {
	case len(magic) >= 2 && magic[0] == 0x1f && magic[1] == 0x8b:
		return gzip.NewReader(br)
	case len(magic) >= 3 && magic[0] == 'B' && magic[1] == 'Z' && magic[2] == 'h':
		return bzip2.NewReader(br), nil
	case len(magic) >= 6 && magic[0] == 0xfd && magic[1] == '7' && magic[2] == 'z' &&
		magic[3] == 'X' && magic[4] == 'Z' && magic[5] == 0x00:
		return xz.NewReader(br)
	case len(magic) >= 4 && magic[0] == 0x28 && magic[1] == 0xb5 && magic[2] == 0x2f && magic[3] == 0xfd:
		zr, zerr := zstd.NewReader(br)
		if zerr != nil {
			return nil, zerr
		}
		return zr.IOReadCloser(), nil
	default:
		return nil, fmt.Errorf("unsupported or missing payload compression")
	}
}

// readHeader reads one RPM header. When pad8 is set, trailing bytes are consumed
// so the next read starts on an 8-byte boundary (as after the signature header).
func readHeader(br *bufio.Reader, pad8 bool) (*Header, int, error) {
	intro := make([]byte, 16)
	if _, err := io.ReadFull(br, intro); err != nil {
		return nil, 0, err
	}
	if intro[0] != headerMagic[0] || intro[1] != headerMagic[1] || intro[2] != headerMagic[2] {
		return nil, 0, fmt.Errorf("bad header magic")
	}
	nindex := int(binary.BigEndian.Uint32(intro[8:12]))
	hsize := int(binary.BigEndian.Uint32(intro[12:16]))

	idx := make([]byte, nindex*16)
	if _, err := io.ReadFull(br, idx); err != nil {
		return nil, 0, err
	}
	store := make([]byte, hsize)
	if _, err := io.ReadFull(br, store); err != nil {
		return nil, 0, err
	}

	h := &Header{entries: make(map[int32]entry, nindex), store: store}
	for i := 0; i < nindex; i++ {
		b := idx[i*16 : i*16+16]
		e := entry{
			tag:    int32(binary.BigEndian.Uint32(b[0:4])),
			typ:    int32(binary.BigEndian.Uint32(b[4:8])),
			offset: int32(binary.BigEndian.Uint32(b[8:12])),
			count:  int32(binary.BigEndian.Uint32(b[12:16])),
		}
		h.entries[e.tag] = e
	}

	total := 16 + nindex*16 + hsize
	if pad8 {
		if rem := total % 8; rem != 0 {
			skip := 8 - rem
			if _, err := br.Discard(skip); err != nil {
				return nil, 0, err
			}
			total += skip
		}
	}
	return h, total, nil
}

// String returns a STRING tag value, or "" when absent.
func (h *Header) String(tag int32) string {
	e, ok := h.entries[tag]
	if !ok || (e.typ != typeString && e.typ != typeI18NString) {
		return ""
	}
	return cstr(h.store, int(e.offset))
}

// StringArray returns a STRING_ARRAY tag value.
func (h *Header) StringArray(tag int32) []string {
	e, ok := h.entries[tag]
	if !ok || (e.typ != typeStringArray && e.typ != typeString && e.typ != typeI18NString) {
		return nil
	}
	out := make([]string, 0, e.count)
	off := int(e.offset)
	for i := int32(0); i < e.count; i++ {
		s := cstr(h.store, off)
		out = append(out, s)
		off += len(s) + 1
	}
	return out
}

// Int32Array returns an INT32 tag value.
func (h *Header) Int32Array(tag int32) []int32 {
	e, ok := h.entries[tag]
	if !ok || (e.typ != typeInt32 && e.typ != typeInt16) {
		return nil
	}
	out := make([]int32, 0, e.count)
	off := int(e.offset)
	for i := int32(0); i < e.count; i++ {
		if e.typ == typeInt32 {
			out = append(out, int32(binary.BigEndian.Uint32(h.store[off:off+4])))
			off += 4
		} else {
			out = append(out, int32(binary.BigEndian.Uint16(h.store[off:off+2])))
			off += 2
		}
	}
	return out
}

// cstr reads a NUL-terminated string from store at offset off.
func cstr(store []byte, off int) string {
	if off < 0 || off >= len(store) {
		return ""
	}
	end := off
	for end < len(store) && store[end] != 0 {
		end++
	}
	return string(store[off:end])
}
