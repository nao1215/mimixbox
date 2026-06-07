// Package tac implements the tac applet: concatenate files (or standard input)
// and write them out with the lines in reverse order.
package tac

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
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

	var b strings.Builder
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			fmt.Fprintf(stdio.Err, "tac: %s\n", command.FileError(name, err))
			firstErr = firstNonNil(firstErr)
			continue
		}
		_, err = io.Copy(&b, r)
		_ = r.Close()
		if err != nil {
			fmt.Fprintf(stdio.Err, "tac: %s\n", command.FileError(name, err))
			firstErr = firstNonNil(firstErr)
		}
	}

	if _, err := io.WriteString(stdio.Out, textproc.Reverse(b.String(), sep)); err != nil {
		return command.Failure(err)
	}
	return firstErr
}

func firstNonNil(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
