// Package rm implements the rm applet: remove files or directories, with the
// common GNU options.
package rm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rm applet.
type Command struct{}

// New returns an rm command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rm" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove file(s) or directory(s)" }

type options struct {
	recursive   bool
	force       bool
	verbose     bool
	dir         bool
	interactive bool
}

// Run executes rm.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err)
	recursive := fs.BoolP("recursive", "r", false, "remove directories and their contents recursively")
	// -R is an alias for -r in GNU rm.
	recursiveUpper := fs.BoolP("Recursive", "R", false, "equivalent to -r")
	force := fs.BoolP("force", "f", false, "ignore nonexistent files and arguments, never prompt")
	verbose := fs.BoolP("verbose", "v", false, "explain what is being done")
	dir := fs.BoolP("dir", "d", false, "remove empty directories")
	interactive := fs.BoolP("interactive", "i", false, "prompt before every removal")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		recursive:   *recursive || *recursiveUpper,
		force:       *force,
		verbose:     *verbose,
		dir:         *dir,
		interactive: *interactive,
	}

	paths := fs.Args()
	if len(paths) == 0 {
		if opts.force {
			return nil
		}
		fmt.Fprintf(stdio.Err, "rm: missing operand\n")
		return command.SilentFailure()
	}

	var failed bool
	in := bufio.NewReader(stdio.In)
	for _, path := range paths {
		if err := remove(stdio, in, path, opts); err != nil {
			fmt.Fprintf(stdio.Err, "rm: %s\n", err.Error())
			failed = true
		}
	}

	if failed {
		return command.SilentFailure()
	}
	return nil
}

// remove deletes a single operand according to opts. It returns an error
// describing the failure (without the "rm:" prefix); the caller prints it and
// keeps going so that the remaining operands are still processed.
func remove(stdio command.IO, in *bufio.Reader, path string, opts options) error {
	info, err := os.Lstat(path)
	if err != nil {
		if opts.force {
			// -f ignores nonexistent files and never reports them.
			return nil
		}
		return fmt.Errorf("can't remove %s: No such file or directory exists", path)
	}

	if info.IsDir() {
		if !opts.recursive {
			// -d allows removing an empty directory without -r.
			if !opts.dir {
				return fmt.Errorf("can't remove %s: It's directory", path)
			}
		}
		if !confirm(stdio, in, path, opts) {
			return nil
		}
		if opts.recursive {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		report(stdio, path, opts)
		return nil
	}

	if !confirm(stdio, in, path, opts) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	report(stdio, path, opts)
	return nil
}

// confirm asks the user before removing path when -i is set. The prompt is
// written to stdio.Err and the answer is read from stdio.In (never os.Stdin),
// so the prompt is testable. Answers starting with "y" (case-insensitive)
// approve the removal; anything else (including EOF) keeps the file.
func confirm(stdio command.IO, in *bufio.Reader, path string, opts options) bool {
	if !opts.interactive {
		return true
	}
	fmt.Fprintf(stdio.Err, "rm: remove '%s'? ", path)
	line, err := in.ReadString('\n')
	answer := strings.ToLower(strings.TrimSpace(line))
	if err != nil && answer == "" {
		return false
	}
	return strings.HasPrefix(answer, "y")
}

// report prints a removal notice when -v is set.
func report(stdio command.IO, path string, opts options) {
	if opts.verbose {
		fmt.Fprintf(stdio.Out, "removed '%s'\n", path)
	}
}
