// Package comm implements the comm applet: compare two sorted files line by
// line, showing lines unique to each file and lines common to both.
package comm

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the comm applet.
type Command struct{}

// New returns a comm command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "comm" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compare two sorted files line by line" }

// Run executes comm.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE1 FILE2", stdio.Err).WithHelp(command.Help{
		Description: "Compare two sorted files line by line and print three columns: lines unique to FILE1, lines " +
			"unique to FILE2, and lines common to both. Use -1, -2, or -3 to suppress the corresponding column.",
		Examples: []command.Example{
			{Command: "comm a.txt b.txt", Explain: "Show lines unique to each file and lines common to both."},
			{Command: "comm -12 a.txt b.txt", Explain: "Print only the lines common to both files."},
			{Command: "comm -3 a.txt b.txt", Explain: "Print only the lines that are unique to one file."},
		},
		ExitStatus: "0  success.\n1  a file could not be read or written.",
	})
	no1 := fs.BoolP("suppress-col1", "1", false, "suppress column 1 (lines unique to FILE1)")
	no2 := fs.BoolP("suppress-col2", "2", false, "suppress column 2 (lines unique to FILE2)")
	no3 := fs.BoolP("suppress-col3", "3", false, "suppress column 3 (lines that appear in both files)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) != 2 {
		_, _ = fmt.Fprintln(stdio.Err, "comm: two file operands are required")
		return command.SilentFailure()
	}

	a, err := readLines(stdio, files[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "comm: %s\n", command.FileError(files[0], err))
		return command.SilentFailure()
	}
	b, err := readLines(stdio, files[1])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "comm: %s\n", command.FileError(files[1], err))
		return command.SilentFailure()
	}

	return c.compare(stdio, a, b, !*no1, !*no2, !*no3)
}

// compare walks the two sorted line slices in merge order and writes each line
// in the column selected by show1/show2/show3.
func (c *Command) compare(stdio command.IO, a, b []string, show1, show2, show3 bool) error {
	// col2 indent skips the active leading columns; col3 skips both.
	pad2, pad3 := "", ""
	if show1 {
		pad2 = "\t"
		pad3 = "\t"
	}
	if show2 {
		pad3 += "\t"
	}

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			if err := emit(stdio, show1, "", a[i]); err != nil {
				return err
			}
			i++
		case a[i] > b[j]:
			if err := emit(stdio, show2, pad2, b[j]); err != nil {
				return err
			}
			j++
		default:
			if err := emit(stdio, show3, pad3, a[i]); err != nil {
				return err
			}
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		if err := emit(stdio, show1, "", a[i]); err != nil {
			return err
		}
	}
	for ; j < len(b); j++ {
		if err := emit(stdio, show2, pad2, b[j]); err != nil {
			return err
		}
	}
	return nil
}

// emit writes line in its column (prefixed with pad) when show is true.
func emit(stdio command.IO, show bool, pad, line string) error {
	if !show {
		return nil
	}
	if _, err := io.WriteString(stdio.Out, pad+line+"\n"); err != nil {
		return command.Failure(err)
	}
	return nil
}

// readLines reads name fully and returns its lines without the line endings.
func readLines(stdio command.IO, name string) ([]string, error) {
	r, err := command.Open(stdio, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	var lines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
