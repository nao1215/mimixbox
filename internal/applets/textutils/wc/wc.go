// Package wc implements the wc applet: count lines, words, characters, bytes
// and the longest line length of files (or standard input).
package wc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
)

// Command is the wc applet.
type Command struct{}

// New returns a wc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "wc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print newline, word, and byte counts for each file" }

type selection struct {
	lines, words, chars, bytes, maxLine bool
}

// fileRow is one line of wc output: the counts and the label to print after
// them (empty for standard input, "total" for the summary line).
type fileRow struct {
	count textproc.Count
	name  string
}

// any reports whether at least one column is selected.
func (s selection) any() bool {
	return s.lines || s.words || s.chars || s.bytes || s.maxLine
}

// orDefault returns the GNU default selection (lines, words, bytes) when no
// column flag was given.
func (s selection) orDefault() selection {
	if s.any() {
		return s
	}
	return selection{lines: true, words: true, bytes: true}
}

// Run executes wc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	lines := fs.BoolP("lines", "l", false, "print the newline counts")
	words := fs.BoolP("words", "w", false, "print the word counts")
	chars := fs.BoolP("chars", "m", false, "print the character counts")
	bytes := fs.BoolP("bytes", "c", false, "print the byte counts")
	maxLine := fs.BoolP("max-line-length", "L", false, "print the maximum display width")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	sel := selection{*lines, *words, *chars, *bytes, *maxLine}.orDefault()
	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var rows []fileRow
	var total textproc.Count
	var firstErr error
	// unknownSize records whether any input's size could not be predetermined
	// (a pipe or a directory). GNU wc widens its columns to 7 in that case.
	unknownSize := false

	for _, name := range files {
		if name == "-" {
			unknownSize = true
		}
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "wc: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		count, err := textproc.CountReader(r)
		_ = r.Close()
		label := name
		if name == "-" {
			label = ""
		}
		if err != nil {
			// A directory opens but cannot be read; GNU still prints a zero
			// row for it and continues.
			_, _ = fmt.Fprintf(stdio.Err, "wc: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			unknownSize = true
			rows = append(rows, fileRow{name: label})
			continue
		}
		rows = append(rows, fileRow{count: count, name: label})
		total = total.Add(count)
	}

	if len(rows) > 1 {
		rows = append(rows, fileRow{count: total, name: "total"})
	}

	columns := len(selectedValues(textproc.Count{}, sel))
	width := fieldWidth(rowsToCounts(rows, sel), columns, unknownSize)
	var b strings.Builder
	for _, rw := range rows {
		b.WriteString(formatRow(rw.count, rw.name, sel, width))
		b.WriteByte('\n')
	}
	if _, err := stdio.Out.Write([]byte(b.String())); err != nil {
		return command.Failure(err)
	}
	return firstErr
}

func rowsToCounts(rows []fileRow, sel selection) []int {
	var nums []int
	for _, rw := range rows {
		nums = append(nums, selectedValues(rw.count, sel)...)
	}
	return nums
}

func selectedValues(c textproc.Count, sel selection) []int {
	var vals []int
	if sel.lines {
		vals = append(vals, c.Lines)
	}
	if sel.words {
		vals = append(vals, c.Words)
	}
	if sel.chars {
		vals = append(vals, c.Runes)
	}
	if sel.bytes {
		vals = append(vals, c.Bytes)
	}
	if sel.maxLine {
		vals = append(vals, c.MaxLineWidth)
	}
	return vals
}

// fieldWidth mirrors GNU wc. The base width is the digit count of the largest
// value (minimum 1). When more than one column is printed and any input's size
// could not be predetermined (a pipe or a directory), the width is raised to a
// minimum of 7 so the columns line up; a single column is never padded.
func fieldWidth(nums []int, columns int, unknownSize bool) int {
	width := 1
	for _, n := range nums {
		if w := len(strconv.Itoa(n)); w > width {
			width = w
		}
	}
	if columns > 1 && unknownSize && width < 7 {
		width = 7
	}
	return width
}

func formatRow(c textproc.Count, name string, sel selection, width int) string {
	fields := make([]string, 0, 5)
	for _, v := range selectedValues(c, sel) {
		fields = append(fields, fmt.Sprintf("%*d", width, v))
	}
	line := strings.Join(fields, " ")
	if name != "" {
		line += " " + name
	}
	return line
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
