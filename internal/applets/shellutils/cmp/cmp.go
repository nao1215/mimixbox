// Package cmp implements the cmp applet: compare two files byte by byte and
// report the offset and line number of the first difference.
package cmp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the cmp applet.
type Command struct{}

// New returns a cmp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cmp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compare two files byte by byte" }

// result is the outcome of comparing two byte streams.
type result struct {
	// equal is true when both streams held exactly the same bytes.
	equal bool
	// eofOn names which side hit EOF first when one stream is a proper prefix
	// of the other: 0 means neither, 1 means the first reader, 2 the second.
	eofOn int
	// byteOffset is the 1-based offset of the first differing byte (only
	// meaningful when there is a real difference, not an EOF).
	byteOffset int64
	// line is the 1-based line number containing the first differing byte.
	line int64
	// a and b are the differing byte values at byteOffset.
	a, b byte
}

// compare reads r1 and r2 byte by byte and reports the first difference, or
// that one stream is a prefix of the other, or that they are equal. It is the
// pure, unit-testable core of cmp: it touches no files, flags, or streams of
// its own.
func compare(r1, r2 io.Reader) (result, error) {
	br1 := bufio.NewReader(r1)
	br2 := bufio.NewReader(r2)

	var offset int64
	var line int64 = 1
	for {
		b1, err1 := br1.ReadByte()
		b2, err2 := br2.ReadByte()

		eof1 := errors.Is(err1, io.EOF)
		eof2 := errors.Is(err2, io.EOF)
		// Surface any non-EOF read error to the caller.
		if err1 != nil && !eof1 {
			return result{}, err1
		}
		if err2 != nil && !eof2 {
			return result{}, err2
		}

		switch {
		case eof1 && eof2:
			return result{equal: true}, nil
		case eof1:
			// r1 ended first: r1 is a prefix of r2.
			return result{eofOn: 1}, nil
		case eof2:
			// r2 ended first: r2 is a prefix of r1.
			return result{eofOn: 2}, nil
		}

		offset++
		if b1 != b2 {
			return result{
				byteOffset: offset,
				line:       line,
				a:          b1,
				b:          b2,
			}, nil
		}
		if b1 == '\n' {
			line++
		}
	}
}

// Run executes cmp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE1 [FILE2]", stdio.Err).WithHelp(command.Help{
		Description: "Compare two files byte by byte and report the byte offset and line number of the first " +
			"difference. When FILE2 is omitted (or either file is '-'), standard input is compared.",
		Examples: []command.Example{
			{Command: "cmp a.txt b.txt", Explain: "Report the first byte at which a.txt and b.txt differ."},
			{Command: "cmp -s a.txt b.txt", Explain: "Compare silently, reporting the result only via the exit status."},
			{Command: "cmp -l a.txt b.txt", Explain: "List the offset and octal values of every differing byte."},
		},
		ExitStatus: "0  inputs are identical.\n1  they differ.\n2  an error occurred.",
	})
	silent := fs.BoolP("silent", "s", false, "suppress all normal output")
	_ = fs.Bool("quiet", false, "suppress all normal output")
	verbose := fs.BoolP("verbose", "l", false, "output byte numbers and differing byte values")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	quiet, _ := fs.GetBool("quiet")
	silentMode := *silent || quiet

	operands := fs.Args()
	if len(operands) < 1 {
		_, _ = fmt.Fprintln(stdio.Err, "cmp: missing operand")
		return &command.ExitError{Code: 2}
	}
	if len(operands) > 2 {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: extra operand '%s'\n", operands[2])
		return &command.ExitError{Code: 2}
	}

	name1 := operands[0]
	name2 := "-"
	if len(operands) == 2 {
		name2 = operands[1]
	}
	if name1 == "-" && name2 == "-" {
		_, _ = fmt.Fprintln(stdio.Err, "cmp: at most one operand may be '-'")
		return &command.ExitError{Code: 2}
	}

	r1, err := command.Open(stdio, name1)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: %s\n", command.FileError(name1, err))
		return &command.ExitError{Code: 2}
	}
	defer func() { _ = r1.Close() }()

	r2, err := command.Open(stdio, name2)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: %s\n", command.FileError(name2, err))
		return &command.ExitError{Code: 2}
	}
	defer func() { _ = r2.Close() }()

	res, err := compare(r1, r2)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: %v\n", err)
		return &command.ExitError{Code: 2}
	}

	if res.equal {
		return nil
	}

	if res.eofOn != 0 {
		if !silentMode {
			shorter := name1
			if res.eofOn == 2 {
				shorter = name2
			}
			_, _ = fmt.Fprintf(stdio.Err, "cmp: EOF on %s\n", shorter)
		}
		return &command.ExitError{Code: 1}
	}

	// Files differ at a concrete byte.
	if !silentMode {
		if *verbose {
			_, _ = fmt.Fprintf(stdio.Out, "%d %o %o\n", res.byteOffset, res.a, res.b)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "%s %s differ: byte %d, line %d\n",
				name1, name2, res.byteOffset, res.line)
		}
	}
	return &command.ExitError{Code: 1}
}
