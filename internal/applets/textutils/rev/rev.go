// Package rev implements the rev applet: reverse the characters of every input
// line, reading from files or standard input.
package rev

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rev applet.
type Command struct{}

// New returns a rev command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rev" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Reverse the order of characters in every line" }

// Run executes rev.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	var firstErr error
	for _, name := range files {
		if err := c.revFile(stdio, name); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "rev: %s\n", command.FileError(name, err))
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// revFile reverses every line read from name and writes it to stdout.
func (c *Command) revFile(stdio command.IO, name string) error {
	r, err := command.Open(stdio, name)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if _, err := io.WriteString(stdio.Out, reverseRunes(sc.Text())+"\n"); err != nil {
			return err
		}
	}
	return sc.Err()
}

// reverseRunes returns s with its runes in reverse order so multi-byte UTF-8
// characters are preserved rather than split.
func reverseRunes(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
