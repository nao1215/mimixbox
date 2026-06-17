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
type Command struct {
	// delimiter separates the columns. GNU comm uses a single tab by default;
	// --output-delimiter=STR replaces it. The merge logic builds the per-column
	// padding from repetitions of this string.
	delimiter string
	// terminator ends each output record: "\n" by default, or "\x00" with
	// --zero-terminated/-z.
	terminator string
}

// New returns a comm command.
func New() *Command { return &Command{delimiter: "\t", terminator: "\n"} }

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
			{Command: "comm --output-delimiter=, a.txt b.txt", Explain: "Separate the columns with a comma."},
		},
		ExitStatus: "0  success.\n1  a file could not be read, written, or was not in sorted order.",
	})
	no1 := fs.BoolP("suppress-col1", "1", false, "suppress column 1 (lines unique to FILE1)")
	no2 := fs.BoolP("suppress-col2", "2", false, "suppress column 2 (lines unique to FILE2)")
	no3 := fs.BoolP("suppress-col3", "3", false, "suppress column 3 (lines that appear in both files)")
	outDelim := fs.String("output-delimiter", "", "separate columns with STR")
	zero := fs.BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")
	checkOrder := fs.Bool("check-order", false, "check that the input is correctly sorted")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if fs.Changed("output-delimiter") {
		if *outDelim == "" {
			_, _ = fmt.Fprintln(stdio.Err, "comm: empty --output-delimiter")
			return command.SilentFailure()
		}
		c.delimiter = *outDelim
	}
	if *zero {
		c.terminator = "\x00"
	}

	files := fs.Args()
	if len(files) != 2 {
		_, _ = fmt.Fprintln(stdio.Err, "comm: two file operands are required")
		return command.SilentFailure()
	}

	a, err := readLines(stdio, files[0], *zero)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "comm: %s\n", command.FileError(files[0], err))
		return command.SilentFailure()
	}
	b, err := readLines(stdio, files[1], *zero)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "comm: %s\n", command.FileError(files[1], err))
		return command.SilentFailure()
	}

	if *checkOrder {
		if !sorted(a) {
			_, _ = fmt.Fprintln(stdio.Err, "comm: file 1 is not in sorted order")
			return command.SilentFailure()
		}
		if !sorted(b) {
			_, _ = fmt.Fprintln(stdio.Err, "comm: file 2 is not in sorted order")
			return command.SilentFailure()
		}
	}

	return c.compare(stdio, a, b, !*no1, !*no2, !*no3)
}

// sorted reports whether lines are in non-decreasing order, matching the order
// comm itself assumes for its merge.
func sorted(lines []string) bool {
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			return false
		}
	}
	return true
}

// compare walks the two sorted line slices in merge order and writes each line
// in the column selected by show1/show2/show3.
func (c *Command) compare(stdio command.IO, a, b []string, show1, show2, show3 bool) error {
	// col2 indent skips the active leading columns; col3 skips both. Each skipped
	// column is represented by one delimiter.
	pad2, pad3 := "", ""
	if show1 {
		pad2 = c.delimiter
		pad3 = c.delimiter
	}
	if show2 {
		pad3 += c.delimiter
	}

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] < b[j]:
			if err := c.emit(stdio, show1, "", a[i]); err != nil {
				return err
			}
			i++
		case a[i] > b[j]:
			if err := c.emit(stdio, show2, pad2, b[j]); err != nil {
				return err
			}
			j++
		default:
			if err := c.emit(stdio, show3, pad3, a[i]); err != nil {
				return err
			}
			i++
			j++
		}
	}
	for ; i < len(a); i++ {
		if err := c.emit(stdio, show1, "", a[i]); err != nil {
			return err
		}
	}
	for ; j < len(b); j++ {
		if err := c.emit(stdio, show2, pad2, b[j]); err != nil {
			return err
		}
	}
	return nil
}

// emit writes line in its column (prefixed with pad) when show is true, ending
// the record with the configured terminator.
func (c *Command) emit(stdio command.IO, show bool, pad, line string) error {
	if !show {
		return nil
	}
	if _, err := io.WriteString(stdio.Out, pad+line+c.terminator); err != nil {
		return command.Failure(err)
	}
	return nil
}

// readLines reads name fully and returns its records without the terminators.
// When zero is true, records are split on NUL instead of newline.
func readLines(stdio command.IO, name string, zero bool) ([]string, error) {
	r, err := command.Open(stdio, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	var lines []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
	if zero {
		sc.Split(scanNUL)
	}
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// scanNUL is a bufio.SplitFunc that splits input on NUL bytes, mirroring
// bufio.ScanLines but for NUL-terminated records.
func scanNUL(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
