// Package readahead implements the readahead applet: preload files into the
// page cache so later access is faster.
package readahead

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the readahead applet.
type Command struct{}

// New returns a readahead command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "readahead" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Preload files into the page cache" }

// preload warms one file into the page cache. It is injected so tests can
// observe which files were requested without depending on kernel cache state.
var preload = func(path string) error {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	// Sequentially reading the file is a portable way to pull its pages into
	// the cache; on Linux the kernel's own read-ahead amplifies it.
	_, err = io.Copy(io.Discard, f)
	return err
}

// Run executes readahead.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Preload each FILE into the kernel page cache by reading it sequentially, so a subsequent " +
			"open-and-read is served from memory. This is advisory and read-only: it never modifies the " +
			"files and has no effect other than warming the cache.",
		Examples: []command.Example{
			{Command: "readahead /usr/bin/* ", Explain: "Warm the cache for every file in /usr/bin."},
		},
		ExitStatus: "0  every file was preloaded.\n1  a file could not be read.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("at least one file operand is required")
	}

	failed := false
	for _, file := range files {
		if err := preload(file); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "readahead: %s\n", command.FileError(file, err))
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}
