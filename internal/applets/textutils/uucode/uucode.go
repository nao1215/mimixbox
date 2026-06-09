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
	})
	useBase64 := fs.BoolP("base64", "m", false, "use base64 encoding")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	var data []byte
	var name string
	switch len(operands) {
	case 1:
		name = operands[0]
		data, err = io.ReadAll(stdio.In)
	case 2:
		name = operands[1]
		data, err = os.ReadFile(operands[0]) //nolint:gosec // user-named file
	default:
		_, _ = fmt.Fprintln(stdio.Err, "uuencode: usage: uuencode [-m] [FILE] NAME")
		return command.SilentFailure()
	}
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "uuencode: %v\n", err)
		return command.SilentFailure()
	}

	if *useBase64 {
		writeBase64(stdio.Out, name, data)
	} else {
		writeUU(stdio.Out, name, data)
	}
	return nil
}

func writeUU(w io.Writer, name string, data []byte) {
	var b strings.Builder
	fmt.Fprintf(&b, "begin 644 %s\n", name)
	for off := 0; off < len(data); off += 45 {
		end := off + 45
		if end > len(data) {
			end = len(data)
		}
		line := data[off:end]
		b.WriteByte(enc(byte(len(line))))
		for i := 0; i < len(line); i += 3 {
			var g [3]byte
			copy(g[:], line[i:])
			b.WriteByte(enc(g[0] >> 2))
			b.WriteByte(enc((g[0]<<4 | g[1]>>4) & 0x3f))
			b.WriteByte(enc((g[1]<<2 | g[2]>>6) & 0x3f))
			b.WriteByte(enc(g[2] & 0x3f))
		}
		b.WriteByte('\n')
	}
	b.WriteByte(enc(0))
	b.WriteString("\nend\n")
	_, _ = io.WriteString(w, b.String())
}

func writeBase64(w io.Writer, name string, data []byte) {
	var b strings.Builder
	fmt.Fprintf(&b, "begin-base64 644 %s\n", name)
	encoded := base64.StdEncoding.EncodeToString(data)
	for off := 0; off < len(encoded); off += 60 {
		end := off + 60
		if end > len(encoded) {
			end = len(encoded)
		}
		b.WriteString(encoded[off:end])
		b.WriteByte('\n')
	}
	b.WriteString("====\n")
	_, _ = io.WriteString(w, b.String())
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
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

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
