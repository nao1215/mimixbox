// Package lzw implements the Unix "compress" (.Z) variant of LZW, compatible
// with the system compress/uncompress and with gzip -d. It is shared by the
// compress and uncompress applets.
//
// The format: a 3-byte header (0x1F, 0x9D, max_bits|0x80 for block mode)
// followed by variable-width LZW codes packed least-significant-bit first, in
// "blocks" of eight codes so that the bit width can grow from 9 up to max_bits.
// Code 256 is the CLEAR code (block mode) and codes are assigned from 257.
package lzw

import (
	"bufio"
	"fmt"
	"io"
)

const (
	magic1    = 0x1f
	magic2    = 0x9d
	blockMode = 0x80
	initBits  = 9
	maxBits   = 16
	clearCode = 256
	firstCode = 257 // first dictionary entry in block mode
)

// Compress reads r and writes the .Z-compressed stream to w.
func Compress(r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)

	if _, err := bw.Write([]byte{magic1, magic2, byte(maxBits | blockMode)}); err != nil {
		return err
	}

	e := &encoder{bw: bw, nBits: initBits, maxCode: (1 << initBits) - 1, free: firstCode}
	e.resetDict()

	first, err := br.ReadByte()
	if err == io.EOF {
		return e.finish()
	}
	if err != nil {
		return err
	}
	ent := int(first)

	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		c := int(b)
		key := (c << maxBits) | ent
		if next, ok := e.dict[key]; ok {
			ent = next
			continue
		}
		if err := e.output(ent); err != nil {
			return err
		}
		if e.free < (1 << maxBits) {
			e.dict[key] = e.free
			e.free++
		} else {
			if err := e.output(clearCode); err != nil {
				return err
			}
			e.resetDict()
		}
		ent = c
	}
	if err := e.output(ent); err != nil {
		return err
	}
	return e.finish()
}

// encoder holds the LZW compression state.
type encoder struct {
	bw      *bufio.Writer
	dict    map[int]int
	nBits   uint
	maxCode int
	free    int

	buf    [maxBits]byte // accumulates one block of nBits codes
	offset uint          // bit offset within buf
	nCodes int           // codes accumulated since the last width change
}

// resetDict clears the dictionary and resets the code width, as a CLEAR does.
func (e *encoder) resetDict() {
	e.dict = make(map[int]int, 1<<maxBits)
	e.nBits = initBits
	e.maxCode = (1 << initBits) - 1
	e.free = firstCode
	e.nCodes = 0
}

// output writes one code using the current width, growing the width on block
// boundaries the way ncompress does.
func (e *encoder) output(code int) error {
	// Pack code at the current bit offset, LSB first (may span up to 3 bytes).
	r := e.offset
	for i := uint(0); i < e.nBits; i++ {
		if code&(1<<i) != 0 {
			e.buf[(r+i)>>3] |= 1 << ((r + i) & 7)
		}
	}
	e.offset += e.nBits
	e.nCodes++

	if e.offset == e.nBits<<3 { // a full block of 8 codes
		if _, err := e.bw.Write(e.buf[:e.nBits]); err != nil {
			return err
		}
		e.clearBuf()
	}

	// After CLEAR, the width resets to initBits (handled by resetDict, but the
	// block must be flushed first).
	if code == clearCode {
		return e.flushBlock()
	}

	// Grow the width when the dictionary has filled the current code space.
	if e.free > e.maxCode && e.nBits < maxBits {
		if err := e.flushBlock(); err != nil {
			return err
		}
		e.nBits++
		e.maxCode = (1 << e.nBits) - 1
	}
	return nil
}

// flushBlock writes any partial block and resets the block counters, which is
// what ncompress does at every width change and clear.
func (e *encoder) flushBlock() error {
	if e.offset > 0 {
		n := (e.offset + 7) >> 3
		if _, err := e.bw.Write(e.buf[:n]); err != nil {
			return err
		}
		e.clearBuf()
	}
	e.nCodes = 0
	return nil
}

func (e *encoder) clearBuf() {
	for i := range e.buf {
		e.buf[i] = 0
	}
	e.offset = 0
}

// finish flushes the last partial block and the writer.
func (e *encoder) finish() error {
	if e.offset > 0 {
		n := (e.offset + 7) >> 3
		if _, err := e.bw.Write(e.buf[:n]); err != nil {
			return err
		}
	}
	return e.bw.Flush()
}

// Decompress reads a .Z stream from r and writes the original bytes to w.
func Decompress(r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)

	hdr := make([]byte, 3)
	if _, err := io.ReadFull(br, hdr); err != nil {
		return fmt.Errorf("not a .Z file")
	}
	if hdr[0] != magic1 || hdr[1] != magic2 {
		return fmt.Errorf("not a .Z file")
	}
	mb := int(hdr[2] & 0x1f)
	block := hdr[2]&blockMode != 0
	if mb < initBits || mb > maxBits {
		return fmt.Errorf("unsupported max bits %d", mb)
	}

	d := &decoder{br: br, bw: bw, maxBits: mb, block: block}
	if err := d.run(); err != nil {
		return err
	}
	return bw.Flush()
}

// decoder holds the LZW decompression state.
type decoder struct {
	br      *bufio.Reader
	bw      *bufio.Writer
	maxBits int
	block   bool

	prefix [1 << maxBits]int
	suffix [1 << maxBits]byte
	stack  []byte
}

// run performs the decompression loop, mirroring the encoder's blocked,
// growing-width packing.
func (d *decoder) run() error {
	nBits := uint(initBits)
	maxCode := func(n uint) int { return (1 << n) - 1 }
	free := firstCode
	if !d.block {
		free = clearCode
	}

	// Block-aligned bit reader state.
	var buf [maxBits]byte
	var size int // bits available in buf
	var bitPos int

	readCode := func() (int, bool, error) {
		if bitPos >= size || free > maxCode(nBits) {
			// Width change or buffer exhausted: refill on a block boundary.
			if free > maxCode(nBits) && nBits < uint(d.maxBits) {
				nBits++
			}
			n, err := io.ReadFull(d.br, buf[:nBits])
			if err == io.EOF || (err == io.ErrUnexpectedEOF && n == 0) {
				return 0, false, nil
			}
			if err != nil && err != io.ErrUnexpectedEOF {
				return 0, false, err
			}
			size = n << 3
			bitPos = 0
			if size == 0 {
				return 0, false, nil
			}
		}
		if bitPos+int(nBits) > size {
			return 0, false, nil
		}
		code := 0
		for i := uint(0); i < nBits; i++ {
			if buf[(bitPos+int(i))>>3]&(1<<uint((bitPos+int(i))&7)) != 0 {
				code |= 1 << i
			}
		}
		bitPos += int(nBits)
		return code, true, nil
	}

	// Initialize single-byte entries.
	for i := 0; i < 256; i++ {
		d.suffix[i] = byte(i)
	}

	resetWidth := func() {
		nBits = initBits
		free = firstCode
		size = 0
		bitPos = 0
	}

	first, ok, err := readCode()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	prevCode := first
	if err := d.bw.WriteByte(byte(first)); err != nil {
		return err
	}
	finChar := byte(first)

	for {
		code, ok, err := readCode()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if d.block && code == clearCode {
			free = firstCode
			resetWidth()
			c, ok, err := readCode()
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			prevCode = c
			finChar = byte(c)
			if err := d.bw.WriteByte(finChar); err != nil {
				return err
			}
			continue
		}

		cur := code
		d.stack = d.stack[:0]
		if code >= free {
			// KwKwK case: emit prevCode's string + its first char.
			d.stack = append(d.stack, finChar)
			cur = prevCode
		}
		for cur >= 256 {
			d.stack = append(d.stack, d.suffix[cur])
			cur = d.prefix[cur]
		}
		finChar = byte(cur)
		d.stack = append(d.stack, finChar)
		for i := len(d.stack) - 1; i >= 0; i-- {
			if err := d.bw.WriteByte(d.stack[i]); err != nil {
				return err
			}
		}

		if free < (1 << d.maxBits) {
			d.prefix[free] = prevCode
			d.suffix[free] = finChar
			free++
		}
		prevCode = code
	}
}
