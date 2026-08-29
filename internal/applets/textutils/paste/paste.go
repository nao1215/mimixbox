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
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Write lines consisting of the sequentially corresponding lines from each FILE, " +
			"separated by TABs, to standard output. With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "paste a.txt b.txt", Explain: "Merge lines of a.txt and b.txt side by side."},
			{Command: "paste -d , a.txt b.txt", Explain: "Use a comma instead of a TAB as the separator."},
			{Command: "paste -s a.txt", Explain: "Join all lines of a.txt onto a single line."},
		},
		ExitStatus: "0  success.\n1  a file could not be opened or read.",
	})
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
//
// The columns are read one line at a time rather than slurped up front, because
// every `-` operand names the SAME standard input. GNU paste gives consecutive
// stdin lines to consecutive columns, so `printf 'a\nb\n' | paste - -` prints
// "a<TAB>b" -- reading each operand to EOF in turn would hand every line to the
// first `-` and leave the rest empty.
func (c *Command) parallel(stdio command.IO, files []string, seps []string) error {
	cols, closers, firstErr := openColumns(stdio, files)
	defer func() {
		for _, cl := range closers {
			_ = cl.Close()
		}
	}()

	for {
		row := make([]string, len(cols))
		any := false
		for i, col := range cols {
			line, ok := col.next()
			if !ok {
				continue
			}
			row[i] = line
			any = true
		}
		// Every column is exhausted: the previous row was the last one.
		if !any {
			break
		}

		var b strings.Builder
		for i, cell := range row {
			if i > 0 {
				b.WriteString(seps[(i-1)%len(seps)])
			}
			b.WriteString(cell)
		}
		b.WriteByte('\n')
		if _, err := io.WriteString(stdio.Out, b.String()); err != nil {
			return command.Failure(err)
		}
	}

	for i, col := range cols {
		if err := col.sc.Err(); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "paste: %s\n", command.FileError(cols[i].name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// column is one paste operand, consumed a line at a time. Two columns that both
// name standard input share one scanner, so a read by either advances the same
// stream.
type column struct {
	name string
	sc   *bufio.Scanner
}

// next returns the column's next line, or false once it is exhausted. A
// bufio.Scanner keeps reporting false after EOF, so a shared stdin scanner
// exhausts every column that reads through it at the same moment.
func (c *column) next() (string, bool) {
	if !c.sc.Scan() {
		return "", false
	}
	return c.sc.Text(), true
}

// openColumns opens every operand once. A file that cannot be opened is
// reported and dropped, leaving the remaining columns to be pasted, and the
// returned error makes the run exit nonzero.
func openColumns(stdio command.IO, files []string) (cols []*column, closers []io.Closer, firstErr error) {
	var stdin *bufio.Scanner
	for _, name := range files {
		if name == "-" {
			if stdin == nil {
				r, err := command.Open(stdio, name)
				if err != nil {
					_, _ = fmt.Fprintf(stdio.Err, "paste: %s\n", command.FileError(name, err))
					if firstErr == nil {
						firstErr = command.SilentFailure()
					}
					continue
				}
				closers = append(closers, r)
				stdin = newLineScanner(r)
			}
			cols = append(cols, &column{name: name, sc: stdin})
			continue
		}

		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "paste: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		closers = append(closers, r)
		cols = append(cols, &column{name: name, sc: newLineScanner(r)})
	}
	return cols, closers, firstErr
}

// newLineScanner returns a scanner configured the way this applet reads every
// input: one line at a time, with the applet-wide long-line cap.
func newLineScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
	return sc
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
	sc.Buffer(make([]byte, 0, 64*1024), command.MaxLineSize)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}
