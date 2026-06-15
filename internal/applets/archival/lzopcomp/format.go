package lzopcomp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/adler32"
	"io"

	lzo "github.com/rasky/go-lzo"
)

// The lzop container format (file extension ".lzo") frames LZO1X-compressed
// data with a fixed magic, a versioned file header guarded by an Adler-32
// header checksum, and a sequence of length-prefixed blocks. Each block stores
// the uncompressed length, the compressed length, the Adler-32 of the
// uncompressed data and (when the block actually shrank) the Adler-32 of the
// compressed data, followed by the payload. A zero uncompressed length marks
// end of stream.
//
// References: lzop source (src/lzop.c, conf.h) and the LZO format notes.

// lzopMagic is the 9-byte signature at the start of every .lzo file.
var lzopMagic = []byte{0x89, 0x4c, 0x5a, 0x4f, 0x00, 0x0d, 0x0a, 0x1a, 0x0a}

const (
	// version is the lzop format version this writer emits (lzop 1.x).
	version uint16 = 0x1030
	// libVersion is the advertised LZO library version.
	libVersion uint16 = 0x2080
	// versionNeeded is the minimum reader version required.
	versionNeeded uint16 = 0x0940
	// methodLZO1X1 selects the LZO1X-1 compression method.
	methodLZO1X1 uint8 = 1
	// level is the advertised compression level for LZO1X-1.
	level uint8 = 5

	// flagAdler32D marks that decompressed-data Adler-32 checksums are present.
	flagAdler32D uint32 = 0x00000001
	// flagAdler32C marks that compressed-data Adler-32 checksums are present.
	flagAdler32C uint32 = 0x00000002

	// maxBlockSize is the uncompressed block size used by the writer (256 KiB),
	// matching lzop's default.
	maxBlockSize = 256 * 1024
	// maxUncompressed bounds a single block's claimed uncompressed length when
	// reading, guarding against corrupt/hostile input (64 MiB).
	maxUncompressed = 64 * 1024 * 1024
)

// errCorrupt reports a malformed lzop stream.
var errCorrupt = errors.New("not in lzop format")

// compressStream reads all of r and writes an lzop (.lzo) stream to w.
func compressStream(r io.Reader, w io.Writer) error {
	if err := writeHeader(w); err != nil {
		return err
	}
	buf := make([]byte, maxBlockSize)
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			if werr := writeBlock(w, buf[:n]); werr != nil {
				return werr
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}
	}
	// Trailing zero uncompressed length marks end of stream.
	return binary.Write(w, binary.BigEndian, uint32(0))
}

// writeHeader writes the magic and the file header (with its checksum).
func writeHeader(w io.Writer) error {
	if _, err := w.Write(lzopMagic); err != nil {
		return err
	}

	var hdr bytes.Buffer
	_ = binary.Write(&hdr, binary.BigEndian, version)
	_ = binary.Write(&hdr, binary.BigEndian, libVersion)
	_ = binary.Write(&hdr, binary.BigEndian, versionNeeded)
	hdr.WriteByte(methodLZO1X1)
	hdr.WriteByte(level)
	_ = binary.Write(&hdr, binary.BigEndian, flagAdler32D|flagAdler32C)
	_ = binary.Write(&hdr, binary.BigEndian, uint32(0)) // mode
	_ = binary.Write(&hdr, binary.BigEndian, uint32(0)) // mtime low
	_ = binary.Write(&hdr, binary.BigEndian, uint32(0)) // mtime high
	hdr.WriteByte(0)                                     // file name length (none)

	// Header checksum covers everything written so far (after the magic).
	sum := adler32.Checksum(hdr.Bytes())
	_ = binary.Write(&hdr, binary.BigEndian, sum)

	_, err := w.Write(hdr.Bytes())
	return err
}

// writeBlock compresses src and writes one lzop block. If compression does not
// shrink the data, the block stores the data uncompressed (compressed length
// equals uncompressed length), exactly as lzop does.
func writeBlock(w io.Writer, src []byte) error {
	comp := lzo.Compress1X(src)
	dAdler := adler32.Checksum(src)

	var payload []byte
	var compLen uint32
	storeCompChecksum := false
	if len(comp) < len(src) {
		payload = comp
		compLen = uint32(len(comp))
		storeCompChecksum = true
	} else {
		// Incompressible: store the original bytes verbatim.
		payload = src
		compLen = uint32(len(src))
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(src))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, compLen); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, dAdler); err != nil {
		return err
	}
	if storeCompChecksum {
		if err := binary.Write(w, binary.BigEndian, adler32.Checksum(payload)); err != nil {
			return err
		}
	}
	_, err := w.Write(payload)
	return err
}

// decompressStream reads an lzop (.lzo) stream from r and writes the original
// bytes to w.
func decompressStream(r io.Reader, w io.Writer) error {
	br := r
	flags, err := readHeader(br)
	if err != nil {
		return err
	}
	hasDChecksum := flags&flagAdler32D != 0
	hasCChecksum := flags&flagAdler32C != 0

	for {
		var dstLen uint32
		if err := binary.Read(br, binary.BigEndian, &dstLen); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if dstLen == 0 {
			return nil // end of stream
		}
		if dstLen > maxUncompressed {
			return errCorrupt
		}

		var srcLen uint32
		if err := binary.Read(br, binary.BigEndian, &srcLen); err != nil {
			return err
		}
		if srcLen == 0 || srcLen > dstLen {
			return errCorrupt
		}

		var dAdler, cAdler uint32
		if hasDChecksum {
			if err := binary.Read(br, binary.BigEndian, &dAdler); err != nil {
				return err
			}
		}
		if hasCChecksum && srcLen < dstLen {
			if err := binary.Read(br, binary.BigEndian, &cAdler); err != nil {
				return err
			}
		}

		payload := make([]byte, srcLen)
		if _, err := io.ReadFull(br, payload); err != nil {
			return err
		}
		if hasCChecksum && srcLen < dstLen && adler32.Checksum(payload) != cAdler {
			return fmt.Errorf("%w: compressed checksum mismatch", errCorrupt)
		}

		var out []byte
		if srcLen < dstLen {
			out, err = lzo.Decompress1X(bytes.NewReader(payload), int(srcLen), int(dstLen))
			if err != nil {
				return fmt.Errorf("%w: %v", errCorrupt, err)
			}
		} else {
			out = payload // stored uncompressed
		}
		if len(out) != int(dstLen) {
			return errCorrupt
		}
		if hasDChecksum && adler32.Checksum(out) != dAdler {
			return fmt.Errorf("%w: data checksum mismatch", errCorrupt)
		}
		if _, err := w.Write(out); err != nil {
			return err
		}
	}
}

// readHeader validates the magic and parses the file header, returning the
// flags field. It verifies the header checksum.
func readHeader(br io.Reader) (uint32, error) {
	magic := make([]byte, len(lzopMagic))
	if _, err := io.ReadFull(br, magic); err != nil {
		return 0, errCorrupt
	}
	if !bytes.Equal(magic, lzopMagic) {
		return 0, errCorrupt
	}

	// The header is variable length (it carries an optional file name), so read
	// the fixed prefix, then the name, then the checksum, accumulating bytes to
	// recompute the Adler-32.
	var acc bytes.Buffer
	read := func(n int) ([]byte, error) {
		b := make([]byte, n)
		if _, err := io.ReadFull(br, b); err != nil {
			return nil, errCorrupt
		}
		acc.Write(b)
		return b, nil
	}

	// version(2) libVersion(2) versionNeeded(2) method(1) level(1) flags(4)
	// mode(4) mtimeLow(4) mtimeHigh(4) = 24 bytes.
	if _, err := read(2 + 2 + 2 + 1 + 1); err != nil {
		return 0, err
	}
	flagsBytes, err := read(4)
	if err != nil {
		return 0, err
	}
	flags := binary.BigEndian.Uint32(flagsBytes)

	if _, err := read(4 + 4 + 4); err != nil { // mode, mtime low/high
		return 0, err
	}

	nameLenB, err := read(1)
	if err != nil {
		return 0, err
	}
	if nameLen := int(nameLenB[0]); nameLen > 0 {
		if _, err := read(nameLen); err != nil {
			return 0, err
		}
	}

	// Header checksum (Adler-32 over everything after the magic).
	want := adler32.Checksum(acc.Bytes())
	var sumBytes [4]byte
	if _, err := io.ReadFull(br, sumBytes[:]); err != nil {
		return 0, errCorrupt
	}
	if binary.BigEndian.Uint32(sumBytes[:]) != want {
		return 0, fmt.Errorf("%w: header checksum mismatch", errCorrupt)
	}
	return flags, nil
}
