// Package rmdir implements the rmdir applet: remove empty directories,
// optionally along with their now-empty ancestors, following GNU coreutils
// behavior.
package rmdir

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unicode"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rmdir applet.
type Command struct{}

// New returns a rmdir command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rmdir" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove directory" }

// Run executes rmdir.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... DIRECTORY...", stdio.Err)
	parents := fs.BoolP("parents", "p", false, "remove DIRECTORY and its ancestors")
	verbose := fs.BoolP("verbose", "v", false, "output a diagnostic for every directory processed")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	dirs := fs.Args()
	if len(dirs) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand\n", c.Name())
		return command.SilentFailure()
	}

	var firstErr error
	for _, dir := range dirs {
		if rerr := c.remove(stdio, dir, *parents, *verbose); rerr != nil {
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// remove deletes dir. With parents set, it also removes the now-empty ancestor
// directories. A failure is reported GNU-style on stderr; the returned error
// only signals that the exit code should be non-zero.
func (c *Command) remove(stdio command.IO, dir string, parents, verbose bool) error {
	if err := c.removeOne(stdio, dir, verbose); err != nil {
		return err
	}
	if !parents {
		return nil
	}
	for cur := filepath.Dir(dir); cur != "." && cur != string(filepath.Separator) && cur != filepath.Dir(cur); cur = filepath.Dir(cur) {
		if err := c.removeOne(stdio, cur, verbose); err != nil {
			return err
		}
	}
	return nil
}

// removeOne removes a single directory with os.Remove, which only succeeds on
// empty directories. A non-empty directory yields the GNU error message.
func (c *Command) removeOne(stdio command.IO, dir string, verbose bool) error {
	if verbose {
		_, _ = fmt.Fprintf(stdio.Out, "%s: removing directory, '%s'\n", c.Name(), dir)
	}
	if err := os.Remove(dir); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: failed to remove '%s': %s\n", c.Name(), dir, reason(err))
		return err
	}
	return nil
}

// reason extracts the underlying cause from a possible *os.PathError so the
// message reads like GNU's (e.g. "Directory not empty") without the operation
// and path that os repeats, capitalizing the first letter to match GNU output.
func reason(err error) string {
	var pe *os.PathError
	msg := err.Error()
	if errors.As(err, &pe) {
		msg = pe.Err.Error()
	}
	if msg == "" {
		return msg
	}
	r := []rune(msg)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
