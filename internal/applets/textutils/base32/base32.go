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
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Base32 encode or decode FILE, or standard input when no FILE (or \"-\") is\n" +
			"given, and write the result to standard output. By default the data is\n" +
			"encoded; use -d to decode instead.",
		Examples: []command.Example{
			{Command: "base32 file.bin", Explain: "encode file.bin to base32"},
			{Command: "base32 -d file.b32", Explain: "decode base32 back to bytes"},
			{Command: "echo hello | base32", Explain: "encode standard input"},
		},
		ExitStatus: "0  success.\n1  the input file could not be read, or the data was not valid base32 while decoding.",
	})
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

	if *decode {
		// Decoding stays buffered so an invalid payload produces no partial
		// output: the result is written only after the whole input validates.
		input, err := io.ReadAll(r)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "base32: %s\n", command.FileError(name, err))
			return command.SilentFailure()
		}
		return c.decode(stdio, input, *ignoreGarbage)
	}
	return c.encode(stdio, name, r, *wrap)
}

// operand returns the single FILE operand, defaulting to "-" (standard input)
// when none is given.
func operand(args []string) string {
	if len(args) == 0 {
		return "-"
	}
	return args[0]
}

// encode streams the base32 encoding of r to stdout, wrapping lines after wrap
// characters. A wrap of 0 disables wrapping. Input is consumed incrementally so
// the whole file is never held in memory (issue #952).
func (c *Command) encode(stdio command.IO, name string, r io.Reader, wrap int) error {
	ww := &wrapWriter{w: stdio.Out, cols: wrap}
	enc := stdbase32.NewEncoder(stdbase32.StdEncoding, ww)
	if _, err := io.Copy(enc, r); err != nil {
		_ = enc.Close()
		_, _ = fmt.Fprintf(stdio.Err, "base32: %s\n", command.FileError(name, err))
		return command.SilentFailure()
	}
	if err := enc.Close(); err != nil {
		return command.Failure(err)
	}
	return ww.finish()
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

// wrapWriter inserts a newline after every cols bytes written to w, letting the
// encoder stream while still wrapping output the way GNU base32 does. A cols of
// 0 (or negative) disables wrapping. finish appends the trailing newline.
type wrapWriter struct {
	w    io.Writer
	cols int
	col  int
}

func (ww *wrapWriter) Write(p []byte) (int, error) {
	if ww.cols <= 0 {
		return ww.w.Write(p)
	}
	total := 0
	for len(p) > 0 {
		if ww.col == ww.cols {
			if _, err := io.WriteString(ww.w, "\n"); err != nil {
				return total, err
			}
			ww.col = 0
		}
		n := ww.cols - ww.col
		if n > len(p) {
			n = len(p)
		}
		m, err := ww.w.Write(p[:n])
		total += m
		ww.col += m
		if err != nil {
			return total, err
		}
		p = p[n:]
	}
	return total, nil
}

// finish writes the single trailing newline that GNU base32 always appends.
func (ww *wrapWriter) finish() error {
	_, err := io.WriteString(ww.w, "\n")
	return err
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
