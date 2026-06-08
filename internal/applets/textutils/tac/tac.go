// Package tac implements the tac applet: concatenate files (or standard input)
// and write them out with the lines in reverse order.
package tac

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tac applet.
type Command struct{}

// New returns a tac command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tac" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the file contents from the end to the beginning" }

// Run executes tac.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	separator := fs.StringP("separator", "s", "\n", "use STRING as the record separator instead of newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	sep := *separator
	if sep == "" {
		sep = "\n"
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	// tac has to read all of its input before it can emit anything (the first
	// line it prints is the input's last record), so unlike a forward filter it
	// is inherently non-streaming. Read each operand in turn into one buffer and
	// then write the records out in reverse, separator-by-separator, straight to
	// the output rather than materializing a second reversed copy of the data.
	var b strings.Builder
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tac: %s\n", command.FileError(name, err))
			firstErr = firstNonNil(firstErr)
			continue
		}
		_, err = io.Copy(&b, r)
		_ = r.Close()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tac: %s\n", command.FileError(name, err))
			firstErr = firstNonNil(firstErr)
		}
	}

	if err := writeReversed(stdio.Out, b.String(), sep); err != nil {
		return command.Failure(err)
	}
	return firstErr
}

// writeReversed splits text into records terminated by sep and writes them to w
// in reverse order, so it never builds a second full-size copy of the input the
// way returning a reversed string would.
func writeReversed(w io.Writer, text, sep string) error {
	if text == "" {
		return nil
	}
	var records []string
	for text != "" {
		i := strings.Index(text, sep)
		if i < 0 {
			records = append(records, text)
			break
		}
		records = append(records, text[:i+len(sep)])
		text = text[i+len(sep):]
	}
	for i := len(records) - 1; i >= 0; i-- {
		if _, err := io.WriteString(w, records[i]); err != nil {
			return err
		}
	}
	return nil
}

func firstNonNil(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
