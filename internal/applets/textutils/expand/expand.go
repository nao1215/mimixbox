// Package expand implements the expand applet: convert tabs to spaces, reading
// files (or standard input) and writing to standard output, with the common GNU
// options.
package expand

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the expand applet.
type Command struct{}

// New returns an expand command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "expand" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Convert TAB to N space (default:N=8)" }

type options struct {
	tabStop int
	initial bool
}

// Run executes expand.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Convert tabs in each FILE to spaces, writing to standard output. " +
			"With no FILE, or when FILE is -, read standard input.",
		Examples: []command.Example{
			{Command: "expand file.txt", Explain: "Convert tabs to spaces using 8-column tab stops."},
			{Command: "expand -t 4 file.txt", Explain: "Use 4-column tab stops instead of 8."},
			{Command: "expand -i file.txt", Explain: "Convert only the leading tabs on each line."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. an input file could not be read).",
	})
	tabs := fs.IntP("tabs", "t", 8, "have tabs N characters apart, not 8")
	initial := fs.BoolP("initial", "i", false, "do not convert tabs after non blanks")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	tabStop := *tabs
	if tabStop <= 0 {
		tabStop = 8
	}
	opts := options{tabStop: tabStop, initial: *initial}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var firstErr error
	for _, name := range files {
		r, openErr := command.Open(stdio, name)
		if openErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "expand: %s\n", command.FileError(name, openErr))
			firstErr = keep(firstErr)
			continue
		}
		copyErr := expand(stdio.Out, r, opts)
		_ = r.Close()
		if copyErr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "expand: %s\n", command.FileError(name, copyErr))
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// expand converts the tabs in r to spaces according to opts and writes the
// result to w. Columns are counted by rune so that a tab advances to the next
// multiple of the tab stop.
func expand(w io.Writer, r io.Reader, opts options) error {
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)
	column := 0
	seenNonBlank := false
	for {
		ch, _, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				return bw.Flush()
			}
			return err
		}
		switch ch {
		case '\t':
			if opts.initial && seenNonBlank {
				if _, werr := bw.WriteRune('\t'); werr != nil {
					return werr
				}
				column++
				continue
			}
			spaces := opts.tabStop - (column % opts.tabStop)
			if _, werr := bw.WriteString(strings.Repeat(" ", spaces)); werr != nil {
				return werr
			}
			column += spaces
		case '\n':
			if _, werr := bw.WriteRune('\n'); werr != nil {
				return werr
			}
			column = 0
			seenNonBlank = false
		default:
			if ch != ' ' {
				seenNonBlank = true
			}
			if _, werr := bw.WriteRune(ch); werr != nil {
				return werr
			}
			column++
		}
	}
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
