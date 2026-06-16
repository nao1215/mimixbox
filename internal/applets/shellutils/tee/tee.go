// Package tee implements the tee applet: read from standard input and write to
// standard output and to each named file, fanning the stream out with the
// common GNU options.
package tee

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tee applet.
type Command struct{}

// New returns a tee command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tee" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Read from standard input and write to standard output and files"
}

// Run executes tee.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Copy standard input to standard output and to each FILE. By default each FILE is " +
			"overwritten; with -a the input is appended instead.",
		Examples: []command.Example{
			{Command: "tee out.txt", Explain: "Write standard input to out.txt and to standard output."},
			{Command: "tee -a log.txt", Explain: "Append standard input to log.txt instead of overwriting it."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file could not be written).",
	})
	appendMode := fs.BoolP("append", "a", false, "append to the given FILEs, do not overwrite")
	// -i/--ignore-interrupts is accepted for GNU compatibility; in this
	// in-memory model there is no SIGINT to ignore, so it is a no-op.
	_ = fs.BoolP("ignore-interrupts", "i", false, "ignore interrupt signals")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return tee(stdio, fs.Args(), *appendMode)
}

// tee copies standard input to standard output and to each named file. A file
// that cannot be opened or written is reported on stderr GNU-style and sets the
// exit status to failure, but does not stop writing to standard output or the
// remaining files.
func tee(stdio command.IO, files []string, appendMode bool) error {
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	writers := []io.Writer{stdio.Out}
	var opened []*os.File
	var failed bool
	for _, name := range files {
		f, err := os.OpenFile(name, flags, 0o644) //nolint:gosec // operating on a user-named file is the whole point
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tee: %s\n", command.FileError(name, err))
			failed = true
			continue
		}
		writers = append(writers, f)
		opened = append(opened, f)
	}

	mw := io.MultiWriter(writers...)
	if _, err := io.Copy(mw, stdio.In); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "tee: %s\n", err)
		failed = true
	}

	for _, f := range opened {
		if cerr := f.Close(); cerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tee: %s\n", command.FileError(f.Name(), cerr))
			failed = true
		}
	}

	if failed {
		return command.SilentFailure()
	}
	return nil
}
