// Package base32 implements the base32 applet: encode or decode data from a
// file (or standard input) to standard output, following GNU base32 semantics.
package base32

import (
	"context"
	stdbase32 "encoding/base32"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the base32 applet.
type Command struct{}

// New returns a base32 command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "base32" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Base32 encode/decode from FILE(or STDIN) to STDOUT"
}

// Run executes base32.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]", stdio.Err)
	decode := fs.BoolP("decode", "d", false, "decode data")
	ignoreGarbage := fs.BoolP("ignore-garbage", "i", false, "when decoding, ignore non-alphabet characters")
	wrap := fs.IntP("wrap", "w", 76, "wrap encoded lines after COLS character (default 76).\nUse 0 to disable line wrapping")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	name := operand(fs.Args())
	r, err := command.Open(stdio, name)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "base32: %s\n", command.FileError(name, err))
		return command.SilentFailure()
	}
	defer func() { _ = r.Close() }()

	input, err := io.ReadAll(r)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "base32: %s\n", command.FileError(name, err))
		return command.SilentFailure()
	}

	if *decode {
		return c.decode(stdio, input, *ignoreGarbage)
	}
	return c.encode(stdio, input, *wrap)
}

// operand returns the single FILE operand, defaulting to "-" (standard input)
// when none is given.
func operand(args []string) string {
	if len(args) == 0 {
		return "-"
	}
	return args[0]
}

// encode writes the base32 encoding of input to stdout, wrapping lines after
// wrap characters. A wrap of 0 disables wrapping.
func (c *Command) encode(stdio command.IO, input []byte, wrap int) error {
	encoded := stdbase32.StdEncoding.EncodeToString(input)
	if _, err := io.WriteString(stdio.Out, wrapLines(encoded, wrap)); err != nil {
		return command.Failure(err)
	}
	return nil
}

// decode writes the base32 decoding of input to stdout.
func (c *Command) decode(stdio command.IO, input []byte, ignoreGarbage bool) error {
	s := string(input)
	if ignoreGarbage {
		s = stripGarbage(s)
	} else {
		s = stripWhitespace(s)
	}

	decoded, err := stdbase32.StdEncoding.DecodeString(s)
	if err != nil {
		_, _ = fmt.Fprintln(stdio.Err, "base32: invalid input")
		return command.SilentFailure()
	}
	if _, err := stdio.Out.Write(decoded); err != nil {
		return command.Failure(err)
	}
	return nil
}

// wrapLines inserts a newline every cols characters and appends a trailing
// newline. A cols of 0 (or negative) produces a single line.
func wrapLines(s string, cols int) string {
	if cols <= 0 {
		return s + "\n"
	}
	var b strings.Builder
	for len(s) > cols {
		b.WriteString(s[:cols])
		b.WriteByte('\n')
		s = s[cols:]
	}
	b.WriteString(s)
	b.WriteByte('\n')
	return b.String()
}

// stripWhitespace removes ASCII whitespace, which is never part of a base32
// payload, so encoded input that spans multiple lines still decodes.
func stripWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '\v', '\f':
			return -1
		}
		return r
	}, s)
}

// stripGarbage removes every character that is not part of the standard base32
// alphabet, implementing -i / --ignore-garbage.
func stripGarbage(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '2' && r <= '7':
			return r
		case r == '=':
			return r
		}
		return -1
	}, s)
}
