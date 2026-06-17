// Package uucode implements the uuencode and uudecode applets. uuencode wraps a
// file (or standard input) in the historical uuencoding (or base64 with -m);
// uudecode reverses either form back to the original bytes and file name.
package uucode

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the uuencode or uudecode applet.
type Command struct{ name string }

// NewUuencode returns the uuencode applet.
func NewUuencode() *Command { return &Command{name: "uuencode"} }

// NewUudecode returns the uudecode applet.
func NewUudecode() *Command { return &Command{name: "uudecode"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.name == "uudecode" {
		return "Decode a uuencoded (or base64) file"
	}
	return "Encode a file for transmission over text channels"
}

// Run dispatches to the encoder or decoder.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	if c.name == "uudecode" {
		return runUudecode(stdio, args)
	}
	return runUuencode(stdio, args)
}

// runUuencode implements: uuencode [-m] [FILE] NAME.
func runUuencode(stdio command.IO, args []string) error {
	fs := command.NewFlagSet("uuencode", "[-m] [FILE] NAME", stdio.Err).WithHelp(command.Help{
		Description: "Read FILE (or standard input) and write it, encoded, to standard output, " +
			"wrapped so it survives text-only channels. NAME is the file name recorded in the " +
			"header for uudecode to recreate. -m uses base64 instead of historical uuencoding.",
		Examples: []command.Example{
			{Command: "uuencode data.bin data.bin > data.uue", Explain: "Encode data.bin."},
			{Command: "uuencode -m img.png img.png", Explain: "Encode using base64."},
		},
		ExitStatus: "0  success.\n1  an error occurred.",
	})
	useBase64 := fs.BoolP("base64", "m", false, "use base64 encoding")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	var src io.Reader
	var name string
	switch len(operands) {
	case 1:
		name = operands[0]
		src = stdio.In
	case 2:
		name = operands[1]
		f, oerr := os.Open(operands[0]) //nolint:gosec // user-named file
		if oerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "uuencode: %v\n", oerr)
			return command.SilentFailure()
		}
		defer func() { _ = f.Close() }()
		src = f
	default:
		_, _ = fmt.Fprintln(stdio.Err, "uuencode: usage: uuencode [-m] [FILE] NAME")
		return command.SilentFailure()
	}

	// The input is consumed incrementally (issue #952): the historical encoder
	// reads it in 45-byte lines and the base64 encoder streams through, so the
	// whole file is never held in memory.
	var werr error
	if *useBase64 {
		werr = writeBase64(stdio.Out, name, src)
	} else {
		werr = writeUU(stdio.Out, name, src)
	}
	if werr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "uuencode: %v\n", werr)
		return command.SilentFailure()
	}
	return nil
}

func writeUU(w io.Writer, name string, r io.Reader) error {
	bw := bufio.NewWriter(w)
	_, _ = fmt.Fprintf(bw, "begin 644 %s\n", name)
	buf := make([]byte, 45)
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			line := buf[:n]
			_ = bw.WriteByte(enc(byte(len(line))))
			for i := 0; i < len(line); i += 3 {
				var g [3]byte
				copy(g[:], line[i:])
				_ = bw.WriteByte(enc(g[0] >> 2))
				_ = bw.WriteByte(enc((g[0]<<4 | g[1]>>4) & 0x3f))
				_ = bw.WriteByte(enc((g[1]<<2 | g[2]>>6) & 0x3f))
				_ = bw.WriteByte(enc(g[2] & 0x3f))
			}
			_ = bw.WriteByte('\n')
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}
	}
	_ = bw.WriteByte(enc(0))
	_, _ = bw.WriteString("\nend\n")
	return bw.Flush()
}

func writeBase64(w io.Writer, name string, r io.Reader) error {
	bw := bufio.NewWriter(w)
	_, _ = fmt.Fprintf(bw, "begin-base64 644 %s\n", name)
	lw := &lineWrapper{w: bw, cols: 60}
	enc := base64.NewEncoder(base64.StdEncoding, lw)
	if _, err := io.Copy(enc, r); err != nil {
		_ = enc.Close()
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	lw.finish()
	_, _ = bw.WriteString("====\n")
	return bw.Flush()
}

// lineWrapper writes a newline after every cols bytes, terminating each full or
// final partial line, so base64 output is wrapped while streaming. finish ends a
// trailing partial line; an empty stream produces no line at all.
type lineWrapper struct {
	w    *bufio.Writer
	cols int
	col  int
}

func (lw *lineWrapper) Write(p []byte) (int, error) {
	for _, b := range p {
		if err := lw.w.WriteByte(b); err != nil {
			return 0, err
		}
		lw.col++
		if lw.col == lw.cols {
			if err := lw.w.WriteByte('\n'); err != nil {
				return 0, err
			}
			lw.col = 0
		}
	}
	return len(p), nil
}

func (lw *lineWrapper) finish() {
	if lw.col > 0 {
		_ = lw.w.WriteByte('\n')
		lw.col = 0
	}
}

// enc encodes a 6-bit value as a uuencode character (0 becomes '`').
func enc(c byte) byte {
	if c == 0 {
		return '`'
	}
	return (c & 0x3f) + 0x20
}

// dec decodes a uuencode character to its 6-bit value.
func dec(c byte) byte { return (c - 0x20) & 0x3f }

// runUudecode implements: uudecode [-o OUTPUT] [FILE].
func runUudecode(stdio command.IO, args []string) error {
	fs := command.NewFlagSet("uudecode", "[-o OUTPUT] [FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Decode a uuencoded or base64 (begin-base64) stream from FILE or standard " +
			"input, writing the result to the file named in its header (or to OUTPUT with -o; " +
			"use -o - for standard output).",
		Examples: []command.Example{
			{Command: "uudecode data.uue", Explain: "Recreate the encoded file."},
			{Command: "uudecode -o - data.uue", Explain: "Write the decoded bytes to standard output."},
		},
		ExitStatus: "0  success.\n1  an error occurred.",
	})
	output := fs.StringP("output-file", "o", "", "write to OUTPUT (- for standard output)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	src := stdio.In
	if rest := fs.Args(); len(rest) > 0 && rest[0] != "-" {
		f, oerr := os.Open(rest[0]) //nolint:gosec // user-named file
		if oerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "uudecode: %s\n", command.FileError(rest[0], oerr))
			return command.SilentFailure()
		}
		defer func() { _ = f.Close() }()
		src = f
	}

	name, data, err := decode(src)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "uudecode: %v\n", err)
		return command.SilentFailure()
	}

	dest := *output
	if dest == "" {
		dest = name
	}
	if dest == "-" {
		_, _ = stdio.Out.Write(data)
		return nil
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil { //nolint:gosec // recreate with the documented default mode
		_, _ = fmt.Fprintf(stdio.Err, "uudecode: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// decode reads a uuencoded or base64 stream and returns the embedded name and
// the decoded bytes.
func decode(r io.Reader) (string, []byte, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)

	name := ""
	base64Mode := false
	started := false
	var out []byte

	for sc.Scan() {
		line := sc.Text()
		if !started {
			switch {
			case strings.HasPrefix(line, "begin-base64 "):
				base64Mode = true
				name = headerName(line, "begin-base64 ")
				started = true
			case strings.HasPrefix(line, "begin "):
				name = headerName(line, "begin ")
				started = true
			}
			continue
		}
		if base64Mode {
			if line == "====" {
				break
			}
			chunk, err := base64.StdEncoding.DecodeString(strings.TrimSpace(line))
			if err != nil {
				return "", nil, err
			}
			out = append(out, chunk...)
			continue
		}
		if line == "end" {
			break
		}
		if line == "" || line == "`" {
			continue
		}
		n := int(dec(line[0]))
		if n == 0 {
			continue
		}
		decoded := decodeUULine(line[1:])
		if n > len(decoded) {
			n = len(decoded)
		}
		out = append(out, decoded[:n]...)
	}
	if err := sc.Err(); err != nil {
		return "", nil, err
	}
	if !started {
		return "", nil, fmt.Errorf("no begin line found")
	}
	return name, out, nil
}

// headerName extracts the file name from a "begin[-base64] MODE NAME" line.
func headerName(line, prefix string) string {
	rest := strings.TrimPrefix(line, prefix)
	fields := strings.SplitN(strings.TrimSpace(rest), " ", 2)
	if len(fields) == 2 {
		return fields[1]
	}
	if _, err := strconv.Atoi(strings.TrimSpace(rest)); err == nil {
		return ""
	}
	return strings.TrimSpace(rest)
}

// decodeUULine decodes the body characters of one uuencoded line to bytes.
func decodeUULine(s string) []byte {
	var out []byte
	for i := 0; i+3 < len(s); i += 4 {
		c0, c1, c2, c3 := dec(s[i]), dec(s[i+1]), dec(s[i+2]), dec(s[i+3])
		out = append(out,
			c0<<2|c1>>4,
			c1<<4|c2>>2,
			c2<<6|c3,
		)
	}
	return out
}
