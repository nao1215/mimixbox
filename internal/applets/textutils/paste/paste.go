// Package paste implements the paste applet: merge corresponding lines of
// files, or (with -s) join each file's lines onto a single line.
package paste

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the paste applet.
type Command struct{}

// New returns a paste command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "paste" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Merge lines of files" }

// Run executes paste.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	delims := fs.StringP("delimiters", "d", "\t", "reuse characters from LIST instead of TABs")
	serial := fs.BoolP("serial", "s", false, "paste one file at a time instead of in parallel")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	seps := delimiters(*delims)

	if *serial {
		return c.serial(stdio, files, seps)
	}
	return c.parallel(stdio, files, seps)
}

// delimiters expands the -d LIST into the cycle of separators paste uses,
// translating the common backslash escapes. An empty list means TAB.
func delimiters(list string) []string {
	if list == "" {
		return []string{"\t"}
	}
	var seps []string
	for i := 0; i < len(list); i++ {
		if list[i] == '\\' && i+1 < len(list) {
			i++
			switch list[i] {
			case 'n':
				seps = append(seps, "\n")
			case 't':
				seps = append(seps, "\t")
			case '\\':
				seps = append(seps, "\\")
			case '0':
				seps = append(seps, "")
			default:
				seps = append(seps, string(list[i]))
			}
			continue
		}
		seps = append(seps, string(list[i]))
	}
	if len(seps) == 0 {
		return []string{"\t"}
	}
	return seps
}

// serial joins every line of each file onto one output line, cycling through
// the separators between lines.
func (c *Command) serial(stdio command.IO, files []string, seps []string) error {
	var firstErr error
	for _, name := range files {
		lines, err := readLines(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "paste: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		var b strings.Builder
		for i, line := range lines {
			if i > 0 {
				b.WriteString(seps[(i-1)%len(seps)])
			}
			b.WriteString(line)
		}
		b.WriteByte('\n')
		if _, err := io.WriteString(stdio.Out, b.String()); err != nil {
			return command.Failure(err)
		}
	}
	return firstErr
}

// parallel merges the i-th line of every file, separated by the delimiter
// cycle, until all files are exhausted.
func (c *Command) parallel(stdio command.IO, files []string, seps []string) error {
	cols := make([][]string, 0, len(files))
	maxLen := 0
	var firstErr error
	for _, name := range files {
		lines, err := readLines(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "paste: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		cols = append(cols, lines)
		if len(lines) > maxLen {
			maxLen = len(lines)
		}
	}

	for row := 0; row < maxLen; row++ {
		var b strings.Builder
		for col, lines := range cols {
			if col > 0 {
				b.WriteString(seps[(col-1)%len(seps)])
			}
			if row < len(lines) {
				b.WriteString(lines[row])
			}
		}
		b.WriteByte('\n')
		if _, err := io.WriteString(stdio.Out, b.String()); err != nil {
			return command.Failure(err)
		}
	}
	return firstErr
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
