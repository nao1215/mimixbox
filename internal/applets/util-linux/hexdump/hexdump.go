// Package hexdump implements the hexdump and hd applets: dump a file (or
// standard input) as hexadecimal. hexdump defaults to two-byte little-endian
// words; hd (and hexdump -C) use the canonical hex+ASCII layout.
package hexdump

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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

	data, err := readInput(stdio, fs.Args())
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	if *skip > 0 {
		if *skip >= int64(len(data)) {
			data = nil
		} else {
			data = data[*skip:]
		}
	}
	if *length >= 0 && *length < len(data) {
		data = data[:*length]
	}

	if c.canonical || *canonical {
		writeCanonical(stdio.Out, *skip, data)
	} else {
		writeTwoByte(stdio.Out, *skip, data)
	}
	return nil
}

// readInput concatenates the named files, or reads standard input when none are
// given (or when "-" is given).
func readInput(stdio command.IO, files []string) ([]byte, error) {
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		return io.ReadAll(stdio.In)
	}
	var out []byte
	for _, f := range files {
		b, err := os.ReadFile(f) //nolint:gosec // user-named file
		if err != nil {
			return nil, errors.New(command.FileError(f, err))
		}
		out = append(out, b...)
	}
	return out, nil
}

// writeCanonical writes the "hexdump -C" / hd layout to w.
func writeCanonical(w io.Writer, base int64, data []byte) {
	var b strings.Builder
	for off := 0; off < len(data); off += 16 {
		end := off + 16
		if end > len(data) {
			end = len(data)
		}
		row := data[off:end]
		fmt.Fprintf(&b, "%08x  ", base+int64(off))
		for i := 0; i < 16; i++ {
			if i == 8 {
				b.WriteByte(' ')
			}
			if i < len(row) {
				fmt.Fprintf(&b, "%02x ", row[i])
			} else {
				b.WriteString("   ")
			}
		}
		b.WriteString(" |")
		for _, ch := range row {
			if ch >= 0x20 && ch <= 0x7e {
				b.WriteByte(ch)
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteString("|\n")
	}
	fmt.Fprintf(&b, "%08x\n", base+int64(len(data)))
	_, _ = io.WriteString(w, b.String())
}

// writeTwoByte writes the default hexdump layout: two-byte little-endian words.
func writeTwoByte(w io.Writer, base int64, data []byte) {
	var b strings.Builder
	for off := 0; off < len(data); off += 16 {
		end := off + 16
		if end > len(data) {
			end = len(data)
		}
		row := data[off:end]
		fmt.Fprintf(&b, "%07x", base+int64(off))
		for i := 0; i < 16; i += 2 {
			switch {
			case i+1 < len(row):
				fmt.Fprintf(&b, " %02x%02x", row[i+1], row[i])
			case i < len(row):
				fmt.Fprintf(&b, " %04x", uint16(row[i])) // a trailing odd byte
			default:
				b.WriteString("     ") // pad missing words to keep 8 columns
			}
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%07x\n", base+int64(len(data)))
	_, _ = io.WriteString(w, b.String())
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
