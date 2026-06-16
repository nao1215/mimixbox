// Package cmp implements the cmp applet: compare two files byte by byte and
// report the offset and line number of the first difference.
package cmp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

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

// compareOptions tunes the byte-by-byte comparison: how many bytes of each
// stream to skip up front, and how many bytes to compare at most.
type compareOptions struct {
	// skip1 and skip2 are the number of leading bytes to discard from r1 and r2
	// before comparison begins (--ignore-initial). The reported byteOffset and
	// line are counted from the first compared byte, matching GNU cmp.
	skip1, skip2 int64
	// limit caps the number of byte pairs compared (--bytes). A value <= 0 means
	// "no limit".
	limit int64
}

// compare reads r1 and r2 byte by byte and reports the first difference, or
// that one stream is a prefix of the other, or that they are equal. It is the
// pure, unit-testable core of cmp: it touches no files, flags, or streams of
// its own. opts controls leading-byte skipping and the comparison length cap.
func compare(r1, r2 io.Reader, opts compareOptions) (result, error) {
	br1 := bufio.NewReader(r1)
	br2 := bufio.NewReader(r2)

	// Discard the ignored prefixes. Hitting EOF inside a skip means that stream
	// has no bytes left to compare, so it behaves like an empty (prefix) stream.
	if err := discard(br1, opts.skip1); err != nil && !errors.Is(err, io.EOF) {
		return result{}, err
	}
	if err := discard(br2, opts.skip2); err != nil && !errors.Is(err, io.EOF) {
		return result{}, err
	}

	var offset int64
	var line int64 = 1
	for {
		if opts.limit > 0 && offset >= opts.limit {
			// Reached the --bytes cap with no difference: treat as equal.
			return result{equal: true}, nil
		}

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

// discard reads and throws away n bytes from r, returning any read error
// (including io.EOF if the stream ends before n bytes are consumed).
func discard(r *bufio.Reader, n int64) error {
	for ; n > 0; n-- {
		if _, err := r.ReadByte(); err != nil {
			return err
		}
	}
	return nil
}

// parseIgnoreInitial parses the --ignore-initial value, which is either a single
// count "N" applied to both files, or "N:M" giving separate counts for FILE1 and
// FILE2. It returns the two skip counts.
func parseIgnoreInitial(s string) (int64, int64, error) {
	if s == "" {
		return 0, 0, nil
	}
	if i := strings.IndexByte(s, ':'); i >= 0 {
		n, err := strconv.ParseInt(s[:i], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		m, err := strconv.ParseInt(s[i+1:], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if n < 0 || m < 0 {
			return 0, 0, fmt.Errorf("invalid --ignore-initial value %q", s)
		}
		return n, m, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	if n < 0 {
		return 0, 0, fmt.Errorf("invalid --ignore-initial value %q", s)
	}
	return n, n, nil
}

// sprintc renders a byte the way GNU cmp does for --print-bytes: the character
// itself when printable, with caret notation for control bytes (^X), "^?" for
// DEL, and an "M-" prefix for bytes with the high bit set. It mirrors GNU's
// sprintc helper; the surrounding format string supplies the column spacing.
func sprintc(b byte) string {
	var sb strings.Builder
	c := b
	if c >= 0x80 {
		sb.WriteString("M-")
		c -= 0x80
	}
	switch {
	case c < 0x20:
		sb.WriteByte('^')
		sb.WriteByte(c + 0x40)
	case c == 0x7f:
		sb.WriteString("^?")
	default:
		sb.WriteByte(c)
	}
	return sb.String()
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
			{Command: "cmp -b a.txt b.txt", Explain: "Also show the differing byte values in the difference message."},
			{Command: "cmp -n 100 a.txt b.txt", Explain: "Compare at most the first 100 bytes of each file."},
			{Command: "cmp -i 4:8 a.txt b.txt", Explain: "Skip 4 bytes of a.txt and 8 of b.txt before comparing."},
		},
		ExitStatus: "0  inputs are identical.\n1  they differ.\n2  an error occurred.",
	})
	silent := fs.BoolP("silent", "s", false, "suppress all normal output")
	_ = fs.Bool("quiet", false, "suppress all normal output")
	verbose := fs.BoolP("verbose", "l", false, "output byte numbers and differing byte values")
	printBytes := fs.BoolP("print-bytes", "b", false, "print differing bytes")
	limit := fs.Int64P("bytes", "n", 0, "compare at most LIMIT bytes")
	ignore := fs.StringP("ignore-initial", "i", "", "skip first N bytes of both inputs (or N:M for each)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	quiet, _ := fs.GetBool("quiet")
	silentMode := *silent || quiet

	skip1, skip2, err := parseIgnoreInitial(*ignore)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: invalid --ignore-initial value '%s'\n", *ignore)
		return &command.ExitError{Code: 2}
	}
	if *limit < 0 {
		_, _ = fmt.Fprintf(stdio.Err, "cmp: invalid --bytes value '%d'\n", *limit)
		return &command.ExitError{Code: 2}
	}

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

	res, err := compare(r1, r2, compareOptions{skip1: skip1, skip2: skip2, limit: *limit})
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
		switch {
		case *verbose:
			_, _ = fmt.Fprintf(stdio.Out, "%d %o %o\n", res.byteOffset, res.a, res.b)
		case *printBytes:
			// GNU: "<f1> <f2> differ: byte N, line L is V1 C1 V2 C2", where V is
			// the octal byte value and C the rendered character.
			_, _ = fmt.Fprintf(stdio.Out, "%s %s differ: byte %d, line %d is %3o %s %3o %s\n",
				name1, name2, res.byteOffset, res.line, res.a, sprintc(res.a), res.b, sprintc(res.b))
		default:
			_, _ = fmt.Fprintf(stdio.Out, "%s %s differ: byte %d, line %d\n",
				name1, name2, res.byteOffset, res.line)
		}
	}
	return &command.ExitError{Code: 1}
}
