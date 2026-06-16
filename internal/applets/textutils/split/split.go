// Package split implements the split applet: break an input file into smaller
// output files by a line count or a byte count.
package split

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the split applet.
type Command struct{}

// New returns a split command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "split" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Split a file into pieces" }

// Run executes split.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [INPUT [PREFIX]]", stdio.Err).WithHelp(command.Help{
		Description: "Split INPUT into fixed-size output files named PREFIXaa, PREFIXab, and so on " +
			"(PREFIX defaults to x). With no INPUT, or when INPUT is -, read standard input.",
		Examples: []command.Example{
			{Command: "split -l 100 big.txt part_", Explain: "Split big.txt into files of 100 lines each, named part_aa, part_ab, ..."},
			{Command: "split -b 1M data.bin", Explain: "Split data.bin into 1 MiB pieces named xaa, xab, ..."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. the input could not be read or an option was invalid).",
	})
	lines := fs.IntP("lines", "l", 1000, "put NUMBER lines per output file")
	byteSpec := fs.StringP("bytes", "b", "", "put SIZE bytes per output file (suffixes K, M allowed)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	input := "-"
	prefix := "x"
	if len(rest) > 0 {
		input = rest[0]
	}
	if len(rest) > 1 {
		prefix = rest[1]
	}

	r, err := command.Open(stdio, input)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "split: %s\n", command.FileError(input, err))
		return command.SilentFailure()
	}
	defer func() { _ = r.Close() }()

	if *byteSpec != "" {
		size, perr := parseSize(*byteSpec)
		if perr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "split: invalid number of bytes: %q\n", *byteSpec)
			return command.SilentFailure()
		}
		return c.byBytes(stdio, r, prefix, size)
	}
	if *lines <= 0 {
		_, _ = fmt.Fprintf(stdio.Err, "split: invalid number of lines: %d\n", *lines)
		return command.SilentFailure()
	}
	return c.byLines(stdio, r, prefix, *lines)
}

// parseSize parses a byte count that may carry a K (1024) or M (1048576) suffix.
func parseSize(s string) (int, error) {
	mult := 1
	switch {
	case strings.HasSuffix(s, "K"):
		mult, s = 1024, strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		mult, s = 1024*1024, strings.TrimSuffix(s, "M")
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid size")
	}
	return n * mult, nil
}

// byLines writes perFile lines into each successive output file.
func (c *Command) byLines(stdio command.IO, r io.Reader, prefix string, perFile int) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	idx, count := 0, 0
	var w *os.File
	closeCur := func() error {
		if w != nil {
			err := w.Close()
			w = nil
			return err
		}
		return nil
	}
	for sc.Scan() {
		if w == nil {
			f, err := create(prefix, idx)
			if err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "split: %v\n", err)
				return command.SilentFailure()
			}
			w = f
			idx++
		}
		if _, err := w.WriteString(sc.Text() + "\n"); err != nil {
			_ = closeCur()
			return command.Failure(err)
		}
		count++
		if count == perFile {
			if err := closeCur(); err != nil {
				return command.Failure(err)
			}
			count = 0
		}
	}
	if err := closeCur(); err != nil {
		return command.Failure(err)
	}
	return sc.Err()
}

// byBytes writes perFile bytes into each successive output file.
func (c *Command) byBytes(stdio command.IO, r io.Reader, prefix string, perFile int) error {
	br := bufio.NewReader(r)
	buf := make([]byte, perFile)
	idx := 0
	for {
		n, err := io.ReadFull(br, buf)
		if n > 0 {
			f, cerr := create(prefix, idx)
			if cerr != nil {
				_, _ = fmt.Fprintf(stdio.Err, "split: %v\n", cerr)
				return command.SilentFailure()
			}
			idx++
			if _, werr := f.Write(buf[:n]); werr != nil {
				_ = f.Close()
				return command.Failure(werr)
			}
			if cerr := f.Close(); cerr != nil {
				return command.Failure(cerr)
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return command.Failure(err)
		}
	}
}

// create opens the idx-th output file, named prefix followed by a two-letter
// suffix (aa, ab, ...).
func create(prefix string, idx int) (*os.File, error) {
	name := prefix + suffix(idx)
	return os.Create(name) //nolint:gosec // writing a user-named output file is the point
}

// suffix returns the two-letter split suffix for the idx-th file.
func suffix(idx int) string {
	return string(rune('a'+idx/26)) + string(rune('a'+idx%26))
}
