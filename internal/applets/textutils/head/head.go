// Package head implements the head applet: print the first part of files (or
// standard input).
package head

import (
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
)

// Command is the head applet.
type Command struct{}

// New returns a head command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "head" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the first NUMBER(default=10) lines" }

// Run executes head.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the first part of each FILE (or standard input when no FILE is given) to standard output. " +
			"By default the first 10 lines are printed; with multiple files, each is preceded by a header.",
		Examples: []command.Example{
			{Command: "head notes.txt", Explain: "Print the first 10 lines of notes.txt."},
			{Command: "head -n 5 notes.txt", Explain: "Print the first 5 lines of notes.txt."},
			{Command: "head -c 100 notes.txt", Explain: "Print the first 100 bytes of notes.txt."},
			{Command: "head -z -n 1 notes.txt", Explain: "Treat NUL as the line delimiter and print the first record."},
		},
		ExitStatus: "0  all input was printed successfully.\n1  a file could not be read.",
	})
	lines := fs.IntP("lines", "n", 10, "print the first NUM lines instead of the first 10")
	bytesN := fs.IntP("bytes", "c", 0, "print the first NUM bytes of each file")
	quiet := fs.BoolP("quiet", "q", false, "never print headers giving file names")
	verbose := fs.BoolP("verbose", "v", false, "always print headers giving file names")
	zeroTerminated := fs.BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	showHeader := (len(files) > 1 || *verbose) && !*quiet
	delim := byte('\n')
	if *zeroTerminated {
		delim = '\x00'
	}

	var firstErr error
	for i, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "head: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		if showHeader {
			writeHeader(stdio.Out, name, i == 0)
		}
		if *bytesN > 0 {
			err = textproc.HeadBytes(stdio.Out, r, *bytesN)
		} else {
			err = textproc.HeadRecords(stdio.Out, r, *lines, delim)
		}
		_ = r.Close()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "head: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

func writeHeader(w io.Writer, name string, first bool) {
	label := name
	if name == "-" {
		label = "standard input"
	}
	if first {
		_, _ = fmt.Fprintf(w, "==> %s <==\n", label)
		return
	}
	_, _ = fmt.Fprintf(w, "\n==> %s <==\n", label)
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
