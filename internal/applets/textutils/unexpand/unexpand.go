// Package unexpand implements the unexpand applet: convert runs of blanks
// (spaces) in files (or standard input) into tabs, the inverse of expand.
package unexpand

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the unexpand applet.
type Command struct{}

// New returns an unexpand command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unexpand" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Convert N space to TAB(default:N=8)" }

type options struct {
	tabStop int
	all     bool
}

// Run executes unexpand.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	tabs := fs.IntP("tabs", "t", 8, "have tabs N characters apart instead of 8 (enables -a)")
	all := fs.BoolP("all", "a", false, "convert all blanks, instead of just initial blanks")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{tabStop: *tabs, all: *all}
	if opts.tabStop <= 0 {
		opts.tabStop = 8
	}
	// Specifying --tabs implies --all, matching GNU unexpand.
	if fs.Changed("tabs") {
		opts.all = true
	}

	return run(stdio, fs.Args(), opts)
}

// run unexpands every operand (defaulting to standard input when there are
// none). A failed open or read is reported on stderr but does not stop the
// remaining files; the returned error only sets the exit code, because its
// message was already printed.
func run(stdio command.IO, files []string, opts options) error {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			fmt.Fprintf(stdio.Err, "unexpand: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		err = unexpand(stdio.Out, r, opts)
		_ = r.Close()
		if err != nil {
			fmt.Fprintf(stdio.Err, "unexpand: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// unexpand reads r line by line and writes the converted text to w.
func unexpand(w io.Writer, r io.Reader, opts options) error {
	br := bufio.NewReader(r)
	bw := bufio.NewWriter(w)
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			newline := ""
			if strings.HasSuffix(line, "\n") {
				newline = "\n"
				line = line[:len(line)-1]
			}
			if _, werr := bw.WriteString(convertLine(line, opts)); werr != nil {
				return werr
			}
			if _, werr := bw.WriteString(newline); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return bw.Flush()
}

// convertLine converts blanks to tabs in a single line (no trailing newline).
// Without -a, only the leading run of blanks is converted; with -a, every run
// of blanks is collapsed onto tab stops.
func convertLine(line string, opts options) string {
	var b strings.Builder
	column := 0         // visual column of the next output character
	pendingBlanks := 0  // number of consecutive spaces seen, not yet emitted
	pendingStart := 0   // column at which the current blank run started
	convertible := true // whether the current blank run may become tabs

	flush := func() {
		if pendingBlanks == 0 {
			return
		}
		if convertible {
			col := pendingStart
			end := pendingStart + pendingBlanks
			// Emit tabs up to each tab stop that lies within the run.
			next := col + (opts.tabStop - col%opts.tabStop)
			for next <= end {
				// A tab only pays off when it spans at least one column; GNU
				// emits a tab even for a single column it can fill.
				b.WriteByte('\t')
				col = next
				next = col + opts.tabStop
			}
			// Remaining columns that do not reach a tab stop stay as spaces.
			for col < end {
				b.WriteByte(' ')
				col++
			}
		} else {
			for i := 0; i < pendingBlanks; i++ {
				b.WriteByte(' ')
			}
		}
		pendingBlanks = 0
	}

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch ch {
		case ' ':
			if pendingBlanks == 0 {
				pendingStart = column
			}
			pendingBlanks++
			column++
		case '\t':
			// A literal tab advances to the next tab stop. Treat it as part of
			// the blank run so adjacent spaces realign correctly.
			if pendingBlanks == 0 {
				pendingStart = column
			}
			next := column + (opts.tabStop - column%opts.tabStop)
			pendingBlanks += next - column
			column = next
		default:
			flush()
			// After the first non-blank, further blank runs are only converted
			// when -a is set.
			if !opts.all {
				convertible = false
			}
			b.WriteByte(ch)
			column++
		}
	}
	flush()
	return b.String()
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
