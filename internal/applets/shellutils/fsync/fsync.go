// Package fsync implements the fsync applet: flush a file's buffered data and
// metadata to the underlying storage with fsync(2).
package fsync

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fsync applet.
type Command struct{}

// New returns an fsync command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fsync" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Flush a file's data to storage with fsync(2)" }

// Run executes fsync.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Flush each FILE's buffered data and metadata to the underlying storage device " +
			"by calling fsync(2) on it. At least one FILE is required.",
		Examples: []command.Example{
			{Command: "fsync data.db", Explain: "Ensure data.db is durably on disk."},
		},
		ExitStatus: "0  every file was synced.\n1  a file could not be opened or synced.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "fsync: missing file operand")
		return command.SilentFailure()
	}

	var failed bool
	for _, name := range files {
		if err := syncFile(name); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fsync: %s\n", command.FileError(name, err))
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

func syncFile(name string) error {
	f, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Sync()
}
