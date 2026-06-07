// Package ln implements the ln applet: create hard or symbolic links between
// files, with the common GNU options.
package ln

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ln applet.
type Command struct{}

// New returns an ln command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ln" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create hard or symbolic link" }

type options struct {
	symbolic bool
	force    bool
	verbose  bool
}

// Run executes ln.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... TARGET [LINK_NAME]", stdio.Err)
	symbolic := fs.BoolP("symbolic", "s", false, "make symbolic links instead of hard links")
	force := fs.BoolP("force", "f", false, "remove existing destination files")
	verbose := fs.BoolP("verbose", "v", false, "print name of each linked file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing file operand\n", c.Name())
		return command.SilentFailure()
	}

	opts := options{symbolic: *symbolic, force: *force, verbose: *verbose}
	return c.run(stdio, operands, opts)
}

// run dispatches to the correct linking form based on the operands.
//
//	ln TARGET             -> link in cwd named like target's base
//	ln TARGET LINK_NAME   -> link TARGET as LINK_NAME
//	ln TARGET... DIRECTORY-> link each TARGET into DIRECTORY
func (c *Command) run(stdio command.IO, operands []string, opts options) error {
	// Single operand: create a link in the current directory whose name is
	// the base name of the target.
	if len(operands) == 1 {
		target := operands[0]
		return c.link(stdio, target, filepath.Base(target), opts)
	}

	last := operands[len(operands)-1]

	// When the final operand is an existing directory, link every target
	// into it (GNU's "ln TARGET... DIRECTORY" form).
	if isDir(last) {
		var firstErr error
		for _, target := range operands[:len(operands)-1] {
			linkName := filepath.Join(last, filepath.Base(target))
			if err := c.link(stdio, target, linkName, opts); err != nil {
				firstErr = keep(firstErr)
			}
		}
		return firstErr
	}

	// Exactly two operands and the second is not a directory: the second is
	// the link name.
	if len(operands) == 2 {
		return c.link(stdio, operands[0], operands[1], opts)
	}

	_, _ = fmt.Fprintf(stdio.Err, "%s: target '%s' is not a directory\n", c.Name(), last)
	return command.SilentFailure()
}

// link creates a single hard or symbolic link from target to linkName,
// honoring -f (force) and -v (verbose).
func (c *Command) link(stdio command.IO, target, linkName string, opts options) error {
	if opts.force {
		if err := removeExisting(linkName); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: cannot remove '%s': %s\n", c.Name(), linkName, reason(err))
			return command.SilentFailure()
		}
	}

	if opts.symbolic {
		if err := os.Symlink(target, linkName); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: failed to create symbolic link '%s': %s\n", c.Name(), linkName, reason(err))
			return command.SilentFailure()
		}
	} else {
		if err := os.Link(target, linkName); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: failed to create hard link '%s': %s\n", c.Name(), linkName, reason(err))
			return command.SilentFailure()
		}
	}

	if opts.verbose {
		if opts.symbolic {
			_, _ = fmt.Fprintf(stdio.Out, "'%s' -> '%s'\n", linkName, target)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "'%s' => '%s'\n", linkName, target)
		}
	}
	return nil
}

// removeExisting deletes path when it exists, ignoring a not-exist condition so
// -f stays quiet about a missing destination.
func removeExisting(path string) error {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Remove(path)
}

// isDir reports whether path is an existing directory, following symlinks.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// reason renders an error GNU-style: "File exists" rather than the verbose
// "symlink target link: file exists" produced by os.LinkError.
func reason(err error) string {
	switch {
	case os.IsExist(err):
		return "File exists"
	case os.IsNotExist(err):
		return "No such file or directory"
	case os.IsPermission(err):
		return "Permission denied"
	}
	var le *os.LinkError
	if errors.As(err, &le) {
		return capitalize(le.Err.Error())
	}
	var pe *os.PathError
	if errors.As(err, &pe) {
		return capitalize(pe.Err.Error())
	}
	return capitalize(err.Error())
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
