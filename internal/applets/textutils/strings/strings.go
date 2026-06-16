// Package strings implements the strings applet: print sequences of printable
// characters of at least a minimum length found in files (or standard input).
package strings

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the strings applet.
type Command struct{}

// New returns a strings command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "strings" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print printable character sequences in files" }

// Run executes strings.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the printable character sequences of at least a minimum length (4 by default) " +
			"found in each FILE. With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "strings program.bin", Explain: "Print the printable strings found in program.bin."},
			{Command: "strings -n 8 program.bin", Explain: "Print only strings of at least eight characters."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be read or -n was invalid).",
	})
	minLen := fs.IntP("bytes", "n", 4, "print sequences of at least N printable characters")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *minLen < 1 {
		_, _ = fmt.Fprintf(stdio.Err, "strings: invalid minimum string length: %d\n", *minLen)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var firstErr error
	for _, name := range files {
		if err := c.scan(stdio, name, *minLen); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "strings: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// scan reads name and prints every run of printable bytes that is at least
// minLen characters long.
func (c *Command) scan(stdio command.IO, name string, minLen int) error {
	r, err := command.Open(stdio, name)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	br := bufio.NewReader(r)
	cur := make([]byte, 0, 64)
	flush := func() error {
		if len(cur) >= minLen {
			if _, err := stdio.Out.Write(append(cur, '\n')); err != nil {
				return err
			}
		}
		cur = cur[:0]
		return nil
	}

	for {
		b, err := br.ReadByte()
		if err != nil {
			if err == io.EOF {
				return flush()
			}
			return err
		}
		if isPrintable(b) {
			cur = append(cur, b)
			continue
		}
		if ferr := flush(); ferr != nil {
			return ferr
		}
	}
}

// isPrintable reports whether b is a printable ASCII character (including the
// space and tab that strings treats as part of a run).
func isPrintable(b byte) bool {
	return b == '\t' || (b >= 0x20 && b <= 0x7e)
}
