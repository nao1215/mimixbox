// Package od implements the od applet: dump files (or standard input) in octal
// and other formats. It mirrors the common behavior of GNU od, including an
// address (offset) column followed by formatted byte groups and a final address
// line giving the total length.
package od

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// bytesPerLine is the number of input bytes dumped on each output line.
const bytesPerLine = 16

// Command is the od applet.
type Command struct{}

// New returns an od command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "od" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Dump files in octal and other formats" }

// Run executes od.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	addr := fs.StringP("address-radix", "A", "o", "decide how file offsets are printed (o, x, d, n)")
	typ := fs.StringP("format", "t", "", "select output format")
	b := fs.BoolP("octal-bytes", "b", false, "same as -t o1")
	cFlag := fs.BoolP("chars", "c", false, "same as -t c")
	x := fs.BoolP("hex-words", "x", false, "same as -t x2")
	o := fs.BoolP("octal-words", "o", false, "same as -t o2")
	d := fs.BoolP("decimal-words", "d", false, "same as -t u2")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	radix, rerr := parseRadix(*addr)
	if rerr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "od: %v\n", rerr)
		return command.Failure(rerr)
	}

	format := selectFormat(*typ, *b, *cFlag, *x, *o, *d)
	ft, ferr := parseType(format)
	if ferr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "od: %v\n", ferr)
		return command.Failure(ferr)
	}

	data, readErr := readAll(stdio, fs.Args())
	if _, werr := io.WriteString(stdio.Out, dump(data, radix, ft)); werr != nil {
		return command.Failure(werr)
	}
	return readErr
}

// selectFormat resolves the effective output type. An explicit -t wins;
// otherwise the convenience flags map to their type strings, defaulting to the
// GNU default of "o2" (2-byte octal words).
func selectFormat(typ string, b, c, x, o, d bool) string {
	switch {
	case typ != "":
		return typ
	case b:
		return "o1"
	case c:
		return "c"
	case x:
		return "x2"
	case o:
		return "o2"
	case d:
		return "u2"
	default:
		return "o2"
	}
}

// readAll reads every operand (defaulting to standard input when there are
// none) and returns the concatenated bytes. A failed open or read is reported
// on stderr but does not stop the remaining files; the returned error only sets
// the exit code, because its message was already printed.
func readAll(stdio command.IO, files []string) ([]byte, error) {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var data []byte
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "od: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		b, err := io.ReadAll(r)
		_ = r.Close()
		data = append(data, b...)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "od: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}
	return data, firstErr
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}

// radix selects how the address (offset) column is rendered.
type radix int

const (
	radixOctal radix = iota
	radixHex
	radixDecimal
	radixNone
)

// parseRadix maps an -A argument to a radix value.
func parseRadix(s string) (radix, error) {
	switch s {
	case "o":
		return radixOctal, nil
	case "x":
		return radixHex, nil
	case "d":
		return radixDecimal, nil
	case "n":
		return radixNone, nil
	default:
		return 0, fmt.Errorf("invalid output address radix '%s'; it must be one character from [doxn]", s)
	}
}

// address renders offset in the given radix, padded to the GNU column width
// (octal/decimal use 7 digits, hex uses 6). When the radix is none the empty
// string is returned.
func address(offset int, r radix) string {
	switch r {
	case radixOctal:
		return fmt.Sprintf("%07o", offset)
	case radixHex:
		return fmt.Sprintf("%06x", offset)
	case radixDecimal:
		return fmt.Sprintf("%07d", offset)
	default: // radixNone
		return ""
	}
}

// formatType describes one -t output type: the size of each unit in bytes and
// how a unit is rendered into a fixed-width field.
type formatType struct {
	unit  int // number of bytes consumed per value
	width int // field width (excluding the leading space separator)
	value func(b []byte) string
}

// parseType maps a -t argument to a formatType. Supported types are o1/o2,
// x1/x2, d1/d2, u1/u2, c and a.
func parseType(s string) (formatType, error) {
	switch s {
	case "o1":
		return formatType{unit: 1, width: 3, value: octalUnit}, nil
	case "o2":
		return formatType{unit: 2, width: 6, value: octalUnit}, nil
	case "x1":
		return formatType{unit: 1, width: 2, value: hexUnit}, nil
	case "x2":
		return formatType{unit: 2, width: 4, value: hexUnit}, nil
	case "d1":
		return formatType{unit: 1, width: 4, value: signedUnit}, nil
	case "d2":
		return formatType{unit: 2, width: 6, value: signedUnit}, nil
	case "u1":
		return formatType{unit: 1, width: 3, value: unsignedUnit}, nil
	case "u2":
		return formatType{unit: 2, width: 5, value: unsignedUnit}, nil
	case "c":
		return formatType{unit: 1, width: 3, value: charUnit}, nil
	case "a":
		return formatType{unit: 1, width: 3, value: namedUnit}, nil
	default:
		return formatType{}, fmt.Errorf("invalid type string '%s'", s)
	}
}

// little assembles up to unit bytes (little-endian) into an unsigned integer,
// treating missing high bytes as zero (matching GNU od on a trailing partial
// unit).
func little(b []byte) uint64 {
	var v uint64
	for i := len(b) - 1; i >= 0; i-- {
		v = v<<8 | uint64(b[i])
	}
	return v
}

func octalUnit(b []byte) string {
	switch len(b) {
	case 1:
		return fmt.Sprintf("%03o", b[0])
	default:
		return fmt.Sprintf("%06o", little(b))
	}
}

func hexUnit(b []byte) string {
	switch len(b) {
	case 1:
		return fmt.Sprintf("%02x", b[0])
	default:
		return fmt.Sprintf("%04x", little(b))
	}
}

func signedUnit(b []byte) string {
	if len(b) == 1 {
		return fmt.Sprintf("%d", int8(b[0]))
	}
	return fmt.Sprintf("%d", int16(little(b)))
}

func unsignedUnit(b []byte) string {
	return fmt.Sprintf("%d", little(b))
}

// cEscapes maps the bytes that -t c renders with a C-style escape.
var cEscapes = map[byte]string{
	0x00: `\0`,
	0x07: `\a`,
	0x08: `\b`,
	0x09: `\t`,
	0x0a: `\n`,
	0x0b: `\v`,
	0x0c: `\f`,
	0x0d: `\r`,
}

// charUnit renders a single byte the way -t c does: a C escape for the well
// known control characters, the character itself when printable, otherwise a
// 3-digit octal value.
func charUnit(b []byte) string {
	c := b[0]
	if esc, ok := cEscapes[c]; ok {
		return esc
	}
	if c >= 0x20 && c < 0x7f {
		return string(c)
	}
	return fmt.Sprintf("%03o", c)
}

// asciiNames holds the named-character spellings used by -t a for the low 128
// bytes (control characters plus space).
var asciiNames = [...]string{
	"nul", "soh", "stx", "etx", "eot", "enq", "ack", "bel",
	"bs", "ht", "nl", "vt", "ff", "cr", "so", "si",
	"dle", "dc1", "dc2", "dc3", "dc4", "nak", "syn", "etb",
	"can", "em", "sub", "esc", "fs", "gs", "rs", "us",
}

// namedUnit renders a single byte the way -t a does: a short name for control
// characters, "sp" for space, "del" for 0x7f, and the printable character
// itself otherwise. The high bit is ignored, as in GNU od.
func namedUnit(b []byte) string {
	c := b[0] & 0x7f
	switch {
	case int(c) < len(asciiNames):
		return asciiNames[c]
	case c == 0x20:
		return "sp"
	case c == 0x7f:
		return "del"
	default:
		return string(c)
	}
}

// dump renders the whole input: one line per bytesPerLine bytes, each line an
// address column (unless the radix is none) followed by space-prefixed,
// right-justified value fields, then a final line giving the total length in
// the address radix.
func dump(data []byte, r radix, ft formatType) string {
	var b strings.Builder
	for off := 0; off < len(data); off += bytesPerLine {
		end := off + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		b.WriteString(line(data[off:end], off, r, ft))
		b.WriteByte('\n')
	}
	if addr := address(len(data), r); addr != "" {
		b.WriteString(addr)
		b.WriteByte('\n')
	}
	return b.String()
}

// line renders a single output line: the address for offset (when shown)
// followed by the formatted units for the (already sliced) chunk of bytes.
func line(chunk []byte, offset int, r radix, ft formatType) string {
	var b strings.Builder
	b.WriteString(address(offset, r))
	for i := 0; i < len(chunk); i += ft.unit {
		end := i + ft.unit
		if end > len(chunk) {
			end = len(chunk)
		}
		b.WriteByte(' ')
		b.WriteString(pad(ft.value(chunk[i:end]), ft.width))
	}
	return b.String()
}

// pad right-justifies s in a field of the given width (no truncation when s is
// already wider).
func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
