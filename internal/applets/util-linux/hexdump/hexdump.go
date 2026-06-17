// Package hexdump implements the hexdump and hd applets: dump a file (or
// standard input) as hexadecimal. hexdump defaults to two-byte little-endian
// words; hd (and hexdump -C) use the canonical hex+ASCII layout.
package hexdump

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the hexdump/hd applet.
type Command struct {
	name      string
	canonical bool // hd forces the -C layout
}

// NewHexdump returns the hexdump applet.
func NewHexdump() *Command { return &Command{name: "hexdump"} }

// NewHd returns the hd applet (hexdump -C).
func NewHd() *Command { return &Command{name: "hd", canonical: true} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.canonical {
		return "Dump a file in canonical hex+ASCII (hexdump -C)"
	}
	return "Display a file in hexadecimal (and other formats)"
}

// Run executes the applet.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(c.help())
	canonical := fs.BoolP("canonical", "C", false, "canonical hex+ASCII display")
	length := fs.IntP("length", "n", -1, "interpret only LENGTH bytes of input")
	skip := fs.Int64P("skip", "s", 0, "skip OFFSET bytes from the beginning")
	_ = fs.BoolP("no-squeezing", "v", true, "display all input data (the default; squeezing is not implemented)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// Open every operand up front so an unreadable file fails before any output
	// is produced, matching the previous read-everything behavior.
	readers, closers, oerr := openInputs(stdio, fs.Args())
	if oerr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), oerr)
		return command.SilentFailure()
	}
	defer func() {
		for _, rc := range closers {
			_ = rc.Close()
		}
	}()

	// Stream the concatenated input rather than buffering it all (issue #952):
	// drop the skipped prefix, then bound the dump to LENGTH bytes if requested.
	src := io.MultiReader(readers...)
	if *skip > 0 {
		if _, err := io.CopyN(io.Discard, src, *skip); err != nil && err != io.EOF {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
	}
	if *length >= 0 {
		src = io.LimitReader(src, int64(*length))
	}

	bw := bufio.NewWriter(stdio.Out)
	d := &rowDumper{w: bw, base: *skip, canonical: c.canonical || *canonical}
	if _, err := io.Copy(d, src); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	d.finish()
	if err := bw.Flush(); err != nil {
		return command.Failure(err)
	}
	return nil
}

// openInputs opens each operand (standard input when none are given or for "-").
// All inputs are opened before any is read so a failed open is reported before
// any output is written.
func openInputs(stdio command.IO, files []string) ([]io.Reader, []io.Closer, error) {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var readers []io.Reader
	var closers []io.Closer
	for _, f := range files {
		if f == "-" {
			readers = append(readers, stdio.In)
			continue
		}
		fh, err := os.Open(f) //nolint:gosec // user-named file
		if err != nil {
			for _, rc := range closers {
				_ = rc.Close()
			}
			return nil, nil, errors.New(command.FileError(f, err))
		}
		readers = append(readers, fh)
		closers = append(closers, fh)
	}
	return readers, closers, nil
}

// rowDumper formats streamed input into 16-byte hexdump rows, holding at most a
// single partial row in memory and emitting the trailing offset line on finish.
type rowDumper struct {
	w         *bufio.Writer
	base      int64
	canonical bool
	buf       [16]byte
	n         int
	off       int // offset (relative to base) of the first byte in buf
}

func (d *rowDumper) Write(p []byte) (int, error) {
	for _, b := range p {
		d.buf[d.n] = b
		d.n++
		if d.n == 16 {
			d.flushRow()
			d.off += 16
			d.n = 0
		}
	}
	return len(p), nil
}

func (d *rowDumper) flushRow() {
	if d.canonical {
		writeCanonicalRow(d.w, d.base+int64(d.off), d.buf[:d.n])
	} else {
		writeTwoByteRow(d.w, d.base+int64(d.off), d.buf[:d.n])
	}
}

func (d *rowDumper) finish() {
	end := d.off + d.n
	if d.n > 0 {
		d.flushRow()
	}
	if d.canonical {
		_, _ = fmt.Fprintf(d.w, "%08x\n", d.base+int64(end))
	} else {
		_, _ = fmt.Fprintf(d.w, "%07x\n", d.base+int64(end))
	}
}

// writeCanonicalRow writes one "hexdump -C" / hd row for the bytes in row, whose
// first byte is at absolute offset addr.
func writeCanonicalRow(b *bufio.Writer, addr int64, row []byte) {
	_, _ = fmt.Fprintf(b, "%08x  ", addr)
	for i := 0; i < 16; i++ {
		if i == 8 {
			_ = b.WriteByte(' ')
		}
		if i < len(row) {
			_, _ = fmt.Fprintf(b, "%02x ", row[i])
		} else {
			_, _ = b.WriteString("   ")
		}
	}
	_, _ = b.WriteString(" |")
	for _, ch := range row {
		if ch >= 0x20 && ch <= 0x7e {
			_ = b.WriteByte(ch)
		} else {
			_ = b.WriteByte('.')
		}
	}
	_, _ = b.WriteString("|\n")
}

// writeTwoByteRow writes one default-layout row (two-byte little-endian words)
// for the bytes in row, whose first byte is at absolute offset addr.
func writeTwoByteRow(b *bufio.Writer, addr int64, row []byte) {
	_, _ = fmt.Fprintf(b, "%07x", addr)
	for i := 0; i < 16; i += 2 {
		switch {
		case i+1 < len(row):
			_, _ = fmt.Fprintf(b, " %02x%02x", row[i+1], row[i])
		case i < len(row):
			_, _ = fmt.Fprintf(b, " %04x", uint16(row[i])) // a trailing odd byte
		default:
			_, _ = b.WriteString("     ") // pad missing words to keep 8 columns
		}
	}
	_ = b.WriteByte('\n')
}

func (c *Command) help() command.Help {
	desc := "Display FILE (or standard input) as hexadecimal. The default format is two-byte " +
		"little-endian words; -C selects the canonical hex+ASCII layout."
	if c.canonical {
		desc = "Display FILE (or standard input) in the canonical hex+ASCII layout (hexdump -C)."
	}
	return command.Help{
		Description: desc,
		Examples: []command.Example{
			{Command: c.Name() + " file.bin", Explain: "Dump the whole file."},
			{Command: c.Name() + " -n 64 file.bin", Explain: "Dump only the first 64 bytes."},
		},
		ExitStatus: "0  the input was dumped successfully.\n1  a file could not be read.",
		Notes: []string{
			"Repeated-line squeezing (the '*' line) is not implemented; all lines are shown.",
		},
	}
}
