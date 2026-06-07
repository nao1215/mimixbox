// Package touch implements the touch applet: update the access and
// modification times of each file to the current time, creating the file if it
// does not yet exist.
package touch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the touch applet.
type Command struct{}

// New returns a touch command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "touch" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Update the access and modification times of each FILE to the current time"
}

type options struct {
	noCreate   bool // -c, --no-create
	accessOnly bool // -a
	modifyOnly bool // -m
}

// Run executes touch.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err)
	noCreate := fs.BoolP("no-create", "c", false, "do not create any files")
	accessOnly := fs.BoolP("access", "a", false, "change only the access time")
	modifyOnly := fs.BoolP("modify", "m", false, "change only the modification time")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(stdio.Err, "touch: missing file operand")
		return command.SilentFailure()
	}

	opts := options{
		noCreate:   *noCreate,
		accessOnly: *accessOnly,
		modifyOnly: *modifyOnly,
	}

	var firstErr error
	for _, file := range files {
		if err := touch(file, opts); err != nil {
			fmt.Fprintln(stdio.Err, "touch: "+err.Error())
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// touch updates the access and modification times of file to the current time.
// When the file does not exist it is created, unless -c/--no-create was given.
// The -a and -m options restrict which of the two timestamps is changed.
func touch(file string, opts options) error {
	path := os.ExpandEnv(file)

	info, statErr := os.Stat(path)
	if errors.Is(statErr, os.ErrNotExist) {
		if opts.noCreate {
			return nil
		}
		f, err := os.Create(path) //nolint:gosec // operating on a user-named file is the whole point
		if err != nil {
			return err
		}
		return f.Close()
	}
	if statErr != nil {
		return statErr
	}

	now := time.Now().Local()
	atime, mtime := now, now
	// -a leaves the modification time untouched; -m leaves the access time
	// untouched. Without either flag both default to now.
	if opts.accessOnly && !opts.modifyOnly {
		mtime = info.ModTime()
	}
	if opts.modifyOnly && !opts.accessOnly {
		// The current access time is not available portably; fall back to the
		// existing modification time so -m leaves the file's other timestamp
		// at a sensible value rather than advancing it to now.
		atime = info.ModTime()
	}
	return os.Chtimes(path, atime, mtime)
}
