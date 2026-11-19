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
		line, newline := strings.CutSuffix(sc.Text(), "\n")
		pieces := foldLine(line, width, spaces)
		for i, piece := range pieces {
			// The pieces a line was cut into end with a newline, and
			// the last ends the way the line did: GNU fold writes no
			// newline it did not read.
			if i < len(pieces)-1 || newline {
				piece += "\n"
			}
			if _, err := io.WriteString(stdio.Out, piece); err != nil {
				return err
			}
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

// foldLine splits a single line into chunks no wider than width. When spaces is
// set it prefers to break after the last blank within the chunk, matching
// "fold -s".
func foldLine(line string, width int, spaces bool) []string {
	if line == "" {
		return []string{""}
	}
	var out []string
	for len(line) > width {
		cut := width
		if spaces {
			if idx := strings.LastIndexAny(line[:width], " \t"); idx > 0 {
				cut = idx + 1
			}
		}
		out = append(out, line[:cut])
		line = line[cut:]
	}
	out = append(out, line)
	return out
}
