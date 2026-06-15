package kbd

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

// buildKeymap creates a valid binary keymap with the given present tables, each
// table's entries set to (tableIndex*1000 + key).
func buildKeymap(t *testing.T, present ...int) []byte {
	t.Helper()
	km := &Keymap{}
	for _, p := range present {
		km.Present[p] = true
		for k := 0; k < NrKeys; k++ {
			km.Tables[p][k] = uint16(p*1000 + k)
		}
	}
	var b bytes.Buffer
	if err := EncodeKeymap(&b, km); err != nil {
		t.Fatalf("EncodeKeymap: %v", err)
	}
	return b.Bytes()
}

func TestKeymapRoundTrip(t *testing.T) {
	t.Parallel()
	data := buildKeymap(t, 0, 1, 4)
	km, err := DecodeKeymap(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeKeymap: %v", err)
	}
	if got := km.PresentTables(); len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 4 {
		t.Fatalf("PresentTables = %v", got)
	}
	if km.Tables[4][10] != 4010 {
		t.Errorf("Tables[4][10] = %d, want 4010", km.Tables[4][10])
	}
	// Re-encode and compare bytes.
	var b bytes.Buffer
	if err := EncodeKeymap(&b, km); err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Equal(b.Bytes(), data) {
		t.Error("round-trip bytes differ")
	}
}

func TestDecodeKeymapBadMagic(t *testing.T) {
	t.Parallel()
	_, err := DecodeKeymap(strings.NewReader("not-a-keymap-at-all-padding-padding"))
	if err == nil || !strings.Contains(err.Error(), "bad magic") {
		t.Fatalf("want bad magic error, got %v", err)
	}
}

func TestDecodeKeymapTruncated(t *testing.T) {
	t.Parallel()
	data := buildKeymap(t, 2)
	// Cut off mid-table.
	if _, err := DecodeKeymap(bytes.NewReader(data[:len(data)-50])); err == nil {
		t.Fatal("expected truncation error")
	}
}

func TestEncodeEndianness(t *testing.T) {
	t.Parallel()
	km := &Keymap{}
	km.Present[0] = true
	km.Tables[0][0] = 0x0102
	var b bytes.Buffer
	if err := EncodeKeymap(&b, km); err != nil {
		t.Fatal(err)
	}
	// magic(7) + flags(256), then first entry little-endian.
	off := len(keymapMagic) + MaxKeymaps
	got := binary.LittleEndian.Uint16(b.Bytes()[off : off+2])
	if got != 0x0102 {
		t.Errorf("entry = %#x, want 0x0102", got)
	}
}
