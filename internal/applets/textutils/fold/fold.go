// Package fold implements the fold applet: wrap each input line to a given
// width, optionally breaking at spaces.
package fold

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fold applet.
type Command struct{}

// New returns a fold command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fold" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Wrap each input line to fit in specified width" }

// Run executes fold.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Wrap each input line from FILE (or standard input when no FILE is given) to fit within WIDTH " +
			"columns, writing the result to standard output. With -s, break lines at spaces where possible.",
		Examples: []command.Example{
			{Command: "fold notes.txt", Explain: "Wrap each line of notes.txt to 80 columns."},
			{Command: "fold -w 40 notes.txt", Explain: "Wrap each line to 40 columns."},
			{Command: "fold -s -w 40 notes.txt", Explain: "Wrap to 40 columns, breaking at spaces so words stay intact."},
		},
		ExitStatus: "0  all input was wrapped successfully.\n1  a file could not be read.",
	})
	width := fs.IntP("width", "w", 80, "use WIDTH columns instead of 80")
	spaces := fs.BoolP("spaces", "s", false, "break at spaces")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *width <= 0 {
		_, _ = fmt.Fprintf(stdio.Err, "fold: invalid number of columns: %d\n", *width)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var firstErr error
	for _, name := range files {
		if err := c.foldFile(stdio, name, *width, *spaces); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fold: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// foldFile wraps every line read from name to width columns.
func (c *Command) foldFile(stdio command.IO, name string, width int, spaces bool) error {
	r, err := command.Open(stdio, name)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
	sc.Split(scanLinesKeepingNewline)
	for sc.Scan() {
		if _, err := io.WriteString(stdio.Out, foldLine(sc.Text(), width, spaces)); err != nil {
			return err
		}
	}
	return sc.Err()
}

// scanLinesKeepingNewline is bufio.ScanLines without the stripping: a token
// is a line with its newline, or the last line without one, and a carriage
// return before the newline stays part of the line.
func scanLinesKeepingNewline(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[:i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// tabWidth is the distance between tab stops fold assumes.
const tabWidth = 8

// advance returns the column after c is written at column, the way GNU fold
// counts: a backspace moves back one, a carriage return goes to the start, a
// tab goes to the next tab stop and any other byte takes one column.
func advance(column int, c byte) int {
	switch c {
	case '\b':
		if column > 0 {
			return column - 1
		}
		return 0
	case '\r':
		return 0
	case '\t':
		return column + tabWidth - column%tabWidth
	default:
		return column + 1
	}
}

// foldLine folds one line, read with its newline if it had one, so that no
// output line runs past width columns. It is GNU fold's algorithm: a byte
// that would end past width starts a new line, unless it is the first of the
// line, which is then written alone; with spaces, the break goes after the
// last blank (space or tab) before that byte when there is one. Columns are
// counted by advance, and a newline is written only where a line is broken
// or where the input had one.
func foldLine(line string, width int, spaces bool) string {
	var out strings.Builder
	pending := make([]byte, 0, len(line))
	column := 0
	for k := 0; k < len(line); k++ {
		c := line[k]
		if c == '\n' {
			out.Write(pending)
			out.WriteByte('\n')
			pending, column = pending[:0], 0
			continue
		}
		for {
			next := advance(column, c)
			if next <= width {
				column = next
				pending = append(pending, c)
				break
			}
			if spaces {
				if end := bytes.LastIndexAny(pending, " \t"); end >= 0 {
					out.Write(pending[:end+1])
					out.WriteByte('\n')
					pending = append(pending[:0], pending[end+1:]...)
					column = 0
					for _, b := range pending {
						column = advance(column, b)
					}
					continue
				}
			}
			if len(pending) == 0 {
				column = next
				pending = append(pending, c)
				break
			}
			out.Write(pending)
			out.WriteByte('\n')
			pending, column = pending[:0], 0
		}
	}
	out.Write(pending)
	return out.String()
}
