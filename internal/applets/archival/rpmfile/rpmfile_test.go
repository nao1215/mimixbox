package rpmfile_test

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/hex"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/nao1215/mimixbox/internal/applets/archival/rpmfile"
	"github.com/ulikunitz/xz"
)

// tagval is one header entry for the fixture builder.
type tagval struct {
	tag   int32
	typ   int32
	data  []byte
	count int32
}

func strTag(tag int32, s string) tagval {
	return tagval{tag: tag, typ: 6, data: append([]byte(s), 0), count: 1}
}

func strArrTag(tag int32, ss []string) tagval {
	var b []byte
	for _, s := range ss {
		b = append(b, []byte(s)...)
		b = append(b, 0)
	}
	return tagval{tag: tag, typ: 8, data: b, count: int32(len(ss))}
}

func int32ArrTag(tag int32, vs []int32) tagval {
	b := make([]byte, 4*len(vs))
	for i, v := range vs {
		binary.BigEndian.PutUint32(b[i*4:], uint32(v))
	}
	return tagval{tag: tag, typ: 4, data: b, count: int32(len(vs))}
}

// buildHeader assembles an RPM header (intro + index + store).
func buildHeader(tags []tagval) []byte {
	var store []byte
	var index []byte
	for _, t := range tags {
		offset := int32(len(store))
		entry := make([]byte, 16)
		binary.BigEndian.PutUint32(entry[0:], uint32(t.tag))
		binary.BigEndian.PutUint32(entry[4:], uint32(t.typ))
		binary.BigEndian.PutUint32(entry[8:], uint32(offset))
		binary.BigEndian.PutUint32(entry[12:], uint32(t.count))
		index = append(index, entry...)
		store = append(store, t.data...)
	}
	intro := make([]byte, 16)
	intro[0], intro[1], intro[2], intro[3] = 0x8e, 0xad, 0xe8, 0x01
	binary.BigEndian.PutUint32(intro[8:], uint32(len(tags)))
	binary.BigEndian.PutUint32(intro[12:], uint32(len(store)))
	out := append(intro, index...)
	return append(out, store...)
}

// buildRPM assembles a complete RPM: lead, (empty) signature header, the main
// header and the payload bytes.
func buildRPM(mainTags []tagval, payload []byte) []byte {
	lead := make([]byte, 96)
	lead[0], lead[1], lead[2], lead[3] = 0xed, 0xab, 0xee, 0xdb

	sig := buildHeader(nil) // 16 bytes, already 8-byte aligned
	main := buildHeader(mainTags)

	out := append([]byte{}, lead...)
	out = append(out, sig...)
	out = append(out, main...)
	return append(out, payload...)
}

func gz(data string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(data))
	_ = w.Close()
	return b.Bytes()
}

func TestParseAndPayloadGzip(t *testing.T) {
	t.Parallel()
	rpmBytes := buildRPM(
		[]tagval{strTag(rpmfile.TagName, "hello")},
		gz("CPIO-PAYLOAD"),
	)
	f, err := rpmfile.Open(bytes.NewReader(rpmBytes))
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	if got := f.Header.String(rpmfile.TagName); got != "hello" {
		t.Errorf("Name = %q, want hello", got)
	}
	p, err := f.Payload()
	if err != nil {
		t.Fatalf("Payload error = %v", err)
	}
	data, _ := io.ReadAll(p)
	if string(data) != "CPIO-PAYLOAD" {
		t.Errorf("payload = %q", data)
	}
}

func TestPayloadXz(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	w, _ := xz.NewWriter(&b)
	_, _ = w.Write([]byte("XZDATA"))
	_ = w.Close()
	f, err := rpmfile.Open(bytes.NewReader(buildRPM(nil, b.Bytes())))
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.Payload()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(p)
	if string(data) != "XZDATA" {
		t.Errorf("payload = %q", data)
	}
}

func TestPayloadZstd(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	w, _ := zstd.NewWriter(&b)
	_, _ = w.Write([]byte("ZSTDDATA"))
	_ = w.Close()
	f, err := rpmfile.Open(bytes.NewReader(buildRPM(nil, b.Bytes())))
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.Payload()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(p)
	if string(data) != "ZSTDDATA" {
		t.Errorf("payload = %q", data)
	}
}

func TestTagsAndFileList(t *testing.T) {
	t.Parallel()
	rpmBytes := buildRPM([]tagval{
		strTag(rpmfile.TagName, "pkg"),
		strTag(rpmfile.TagVersion, "1.0"),
		strTag(rpmfile.TagRelease, "2"),
		strTag(rpmfile.TagArch, "x86_64"),
		strArrTag(rpmfile.TagDirnames, []string{"/usr/bin/", "/etc/"}),
		strArrTag(rpmfile.TagBasenames, []string{"tool", "tool.conf"}),
		int32ArrTag(rpmfile.TagDirindexes, []int32{0, 1}),
	}, gz("x"))

	f, err := rpmfile.Open(bytes.NewReader(rpmBytes))
	if err != nil {
		t.Fatal(err)
	}
	h := f.Header
	if h.String(rpmfile.TagVersion) != "1.0" || h.String(rpmfile.TagRelease) != "2" {
		t.Errorf("version/release wrong: %q %q", h.String(rpmfile.TagVersion), h.String(rpmfile.TagRelease))
	}
	dirs := h.StringArray(rpmfile.TagDirnames)
	if len(dirs) != 2 || dirs[0] != "/usr/bin/" {
		t.Errorf("dirnames = %v", dirs)
	}
	idx := h.Int32Array(rpmfile.TagDirindexes)
	if len(idx) != 2 || idx[1] != 1 {
		t.Errorf("dirindexes = %v", idx)
	}
}

func TestMissingTag(t *testing.T) {
	t.Parallel()
	f, err := rpmfile.Open(bytes.NewReader(buildRPM(nil, gz("x"))))
	if err != nil {
		t.Fatal(err)
	}
	if f.Header.String(rpmfile.TagName) != "" {
		t.Error("absent tag should yield empty string")
	}
	if f.Header.StringArray(rpmfile.TagBasenames) != nil {
		t.Error("absent array tag should yield nil")
	}
}

func TestPayloadBzip2(t *testing.T) {
	t.Parallel()
	// Precomputed "bzip2 -c" of "hello bunzip2\n" (Go has no bzip2 compressor).
	const bz2Hex = "425a6839314159265359d77cc601000002d9800010400010001265c21020002200034201a005cd7a817a13b878bb9229c28486bbe63008"
	b, err := hex.DecodeString(bz2Hex)
	if err != nil {
		t.Fatal(err)
	}
	f, err := rpmfile.Open(bytes.NewReader(buildRPM(nil, b)))
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.Payload()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(p)
	if string(data) != "hello bunzip2\n" {
		t.Errorf("payload = %q", data)
	}
}

func TestInt16Array(t *testing.T) {
	t.Parallel()
	// An INT16 array (type 3): two 2-byte big-endian values.
	data := []byte{0x00, 0x05, 0x00, 0x07}
	tv := tagval{tag: rpmfile.TagDirindexes, typ: 3, data: data, count: 2}
	f, err := rpmfile.Open(bytes.NewReader(buildRPM([]tagval{tv}, gz("x"))))
	if err != nil {
		t.Fatal(err)
	}
	got := f.Header.Int32Array(rpmfile.TagDirindexes)
	if len(got) != 2 || got[0] != 5 || got[1] != 7 {
		t.Errorf("int16 array = %v, want [5 7]", got)
	}
}

func TestPaddedSignatureHeader(t *testing.T) {
	t.Parallel()
	// Build an RPM whose signature header is not 8-byte aligned, exercising the
	// padding-skip path in readHeader.
	lead := make([]byte, 96)
	lead[0], lead[1], lead[2], lead[3] = 0xed, 0xab, 0xee, 0xdb
	sig := buildHeader([]tagval{strTag(1, "ab")}) // 16 + 16 + 3 = 35 bytes -> needs 5 pad
	sig = append(sig, make([]byte, (8-len(sig)%8)%8)...)
	main := buildHeader([]tagval{strTag(rpmfile.TagName, "padded")})
	out := append(append(append(append([]byte{}, lead...), sig...), main...), gz("x")...)

	f, err := rpmfile.Open(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	if f.Header.String(rpmfile.TagName) != "padded" {
		t.Errorf("Name = %q, want padded", f.Header.String(rpmfile.TagName))
	}
}

func TestNotAnRPM(t *testing.T) {
	t.Parallel()
	_, err := rpmfile.Open(bytes.NewReader(make([]byte, 200)))
	if err == nil {
		t.Error("expected error for non-RPM data")
	}
}

func TestShortFile(t *testing.T) {
	t.Parallel()
	_, err := rpmfile.Open(bytes.NewReader([]byte{0xed, 0xab}))
	if err == nil {
		t.Error("expected error for truncated lead")
	}
}

func TestUnsupportedPayload(t *testing.T) {
	t.Parallel()
	// A payload with no recognizable compression magic.
	f, err := rpmfile.Open(bytes.NewReader(buildRPM(nil, []byte("plain-bytes"))))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Payload(); err == nil {
		t.Error("expected error for unsupported payload compression")
	}
}
