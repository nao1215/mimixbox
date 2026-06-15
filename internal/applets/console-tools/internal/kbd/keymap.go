// Package kbd holds the binary parsers and serializers shared by the keyboard
// and console-font applets (dumpkmap, loadkmap, loadfont, setfont). Keeping the
// format logic here, separate from any device or ioctl access, lets it be unit
// tested entirely in memory.
package kbd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MaxKeymaps is the number of keymap tables (modifier combinations) the binary
// keymap format can describe, and NrKeys is the number of keys in each table.
// These match the Linux keyboard driver and BusyBox's binary keymap layout.
const (
	MaxKeymaps = 256
	NrKeys     = 256
)

// keymapMagic is the 7-byte signature at the start of a BusyBox binary keymap.
var keymapMagic = []byte("bkeymap")

// Keymap is a decoded binary keymap: a flag for each of the 256 possible
// modifier tables saying whether it is present, and, for each present table, its
// 256 key entries.
type Keymap struct {
	// Present[i] reports whether table i is included.
	Present [MaxKeymaps]bool
	// Tables[i] holds the 256 entries for table i; it is only meaningful when
	// Present[i] is true.
	Tables [MaxKeymaps][NrKeys]uint16
}

// DecodeKeymap parses the BusyBox binary keymap format from r:
//
//	"bkeymap"                 7-byte magic
//	flags[256]                one byte per table, non-zero => table present
//	for each present table:   256 little-endian uint16 key entries
//
// It returns a descriptive error on a bad magic or a truncated stream.
func DecodeKeymap(r io.Reader) (*Keymap, error) {
	br := bufferedAll(r)
	magic := make([]byte, len(keymapMagic))
	if _, err := io.ReadFull(br, magic); err != nil {
		return nil, fmt.Errorf("reading keymap magic: %w", err)
	}
	if !bytes.Equal(magic, keymapMagic) {
		return nil, fmt.Errorf("not a binary keymap: bad magic %q", magic)
	}

	km := &Keymap{}
	flags := make([]byte, MaxKeymaps)
	if _, err := io.ReadFull(br, flags); err != nil {
		return nil, fmt.Errorf("reading keymap table flags: %w", err)
	}
	for i, f := range flags {
		if f == 0 {
			continue
		}
		km.Present[i] = true
		for k := 0; k < NrKeys; k++ {
			var v uint16
			if err := binary.Read(br, binary.LittleEndian, &v); err != nil {
				return nil, fmt.Errorf("reading table %d entry %d: %w", i, k, err)
			}
			km.Tables[i][k] = v
		}
	}
	return km, nil
}

// EncodeKeymap writes km back out in the same binary format DecodeKeymap reads,
// so Decode(Encode(km)) round-trips.
func EncodeKeymap(w io.Writer, km *Keymap) error {
	if _, err := w.Write(keymapMagic); err != nil {
		return err
	}
	var flags [MaxKeymaps]byte
	for i := range km.Present {
		if km.Present[i] {
			flags[i] = 1
		}
	}
	if _, err := w.Write(flags[:]); err != nil {
		return err
	}
	for i := range km.Present {
		if !km.Present[i] {
			continue
		}
		for k := 0; k < NrKeys; k++ {
			if err := binary.Write(w, binary.LittleEndian, km.Tables[i][k]); err != nil {
				return err
			}
		}
	}
	return nil
}

// PresentTables returns the indices of the tables marked present, in order.
func (km *Keymap) PresentTables() []int {
	var idx []int
	for i := range km.Present {
		if km.Present[i] {
			idx = append(idx, i)
		}
	}
	return idx
}

// ErrEmptyKeymap is returned by DecodeKeymap callers that require at least one
// table; it is exported so applets can recognize it.
var ErrEmptyKeymap = errors.New("keymap contains no tables")

// bufferedAll returns r as-is; it exists as a seam in case future formats need
// look-ahead. Kept tiny so the parser reads straight from the stream.
func bufferedAll(r io.Reader) io.Reader { return r }
