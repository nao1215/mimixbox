// Package xxd implements the xxd applet: make a hex dump of its input, or with
// -r convert a hex dump back into binary.
package xxd

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the xxd applet.
type Command struct{}

// New returns an xxd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "xxd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Make a hex dump or do the reverse" }

const bytesPerLine = 16

// Run executes xxd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]", stdio.Err)
	reverse := fs.BoolP("revert", "r", false, "reverse operation: convert a hex dump into binary")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name := operand(fs.Args())
	r, err := command.Open(stdio, name)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "xxd: %s\n", command.FileError(name, err))
		return command.SilentFailure()
	}
	defer func() { _ = r.Close() }()

	if *reverse {
		return c.revert(stdio, r)
	}
	return c.dump(stdio, r)
}

// operand returns the single FILE operand, defaulting to "-" (standard input).
func operand(args []string) string {
	if len(args) == 0 {
		return "-"
	}
	return args[0]
}

// dump writes the canonical xxd hex dump of r to stdout.
func (c *Command) dump(stdio command.IO, r io.Reader) error {
	br := bufio.NewReader(r)
	buf := make([]byte, bytesPerLine)
	offset := 0
	for {
		n, err := io.ReadFull(br, buf)
		if n > 0 {
			if _, werr := io.WriteString(stdio.Out, formatLine(offset, buf[:n])); werr != nil {
				return command.Failure(werr)
			}
			offset += n
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return command.Failure(err)
		}
	}
}

// formatLine renders one dump line: an 8-digit offset, the bytes as
// space-separated 2-byte groups padded to a fixed width, and the printable
// rendering of the bytes.
func formatLine(offset int, data []byte) string {
	var hexPart strings.Builder
	var asciiPart strings.Builder
	for i, b := range data {
		if i > 0 && i%2 == 0 {
			hexPart.WriteByte(' ')
		}
		fmt.Fprintf(&hexPart, "%02x", b)
		if b >= 0x20 && b <= 0x7e {
			asciiPart.WriteByte(b)
		} else {
			asciiPart.WriteByte('.')
		}
	}
	// 16 bytes -> 8 groups of 4 hex chars + 7 separators = 39 columns.
	return fmt.Sprintf("%08x: %-39s  %s\n", offset, hexPart.String(), asciiPart.String())
}

// revert reads an xxd dump and writes the original bytes. The hex column is the
// text between ": " and the two-space gap that precedes the ASCII column.
func (c *Command) revert(stdio command.IO, r io.Reader) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		colon := strings.Index(line, ": ")
		if colon < 0 {
			continue
		}
		rest := line[colon+2:]
		if gap := strings.Index(rest, "  "); gap >= 0 {
			rest = rest[:gap]
		}
		raw := strings.ReplaceAll(rest, " ", "")
		decoded, err := hex.DecodeString(raw)
		if err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "xxd: invalid hex dump")
			return command.SilentFailure()
		}
		if _, err := stdio.Out.Write(decoded); err != nil {
			return command.Failure(err)
		}
	}
	if err := sc.Err(); err != nil {
		return command.Failure(err)
	}
	return nil
}
