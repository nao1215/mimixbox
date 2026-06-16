// Package rm implements the rm applet: remove files or directories, with the
// common GNU options.
package rm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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
	recursive     bool
	force         bool
	verbose       bool
	dir           bool
	interactive   bool
	preserveRoot  bool // refuse to recurse on "/" (GNU default ON)
	oneFileSystem bool // skip subdirectories on a different filesystem
}

// Run executes rm.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Remove each FILE. By default rm does not remove directories; use -r (or -R) to " +
			"remove a directory and its contents recursively, or -d to remove an empty directory. " +
			"With -f nonexistent operands are ignored and no prompt is shown, while -i prompts " +
			"before every removal.",
		Examples: []command.Example{
			{Command: "rm file.txt", Explain: "Remove a single file."},
			{Command: "rm -r dir", Explain: "Remove a directory and everything under it."},
			{Command: "rm -f *.tmp", Explain: "Remove matching files, ignoring any that do not exist."},
			{Command: "rm -i note.txt", Explain: "Prompt for confirmation before removing note.txt."},
		},
		ExitStatus: "0  all operands were removed (or ignored with -f).\n1  a file could not be removed or no operand was given.",
	})
	recursive := fs.BoolP("recursive", "r", false, "remove directories and their contents recursively")
	// -R is an alias for -r in GNU rm.
	recursiveUpper := fs.BoolP("Recursive", "R", false, "equivalent to -r")
	force := fs.BoolP("force", "f", false, "ignore nonexistent files and arguments, never prompt")
	verbose := fs.BoolP("verbose", "v", false, "explain what is being done")
	dir := fs.BoolP("dir", "d", false, "remove empty directories")
	interactive := fs.BoolP("interactive", "i", false, "prompt before every removal")
	// --preserve-root is the GNU default: refuse to recursively operate on "/".
	preserveRoot := fs.Bool("preserve-root", true, "do not remove '/' recursively (default)")
	noPreserveRoot := fs.Bool("no-preserve-root", false, "do not treat '/' specially")
	oneFileSystem := fs.Bool("one-file-system", false, "when removing a hierarchy recursively, skip any directory that is on a file system different from that of the corresponding command line argument")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		recursive:     *recursive || *recursiveUpper,
		force:         *force,
		verbose:       *verbose,
		dir:           *dir,
		interactive:   *interactive,
		preserveRoot:  *preserveRoot && !*noPreserveRoot,
		oneFileSystem: *oneFileSystem,
	}

	paths := fs.Args()
	if len(paths) == 0 {
		if opts.force {
			return nil
		}
		_, _ = fmt.Fprintf(stdio.Err, "rm: missing operand\n")
		return command.SilentFailure()
	}

	var failed bool
	in := bufio.NewReader(stdio.In)
	for _, path := range paths {
		if err := remove(stdio, in, path, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "rm: %s\n", err.Error())
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
		// --preserve-root: refuse to recursively operate on "/".
		if opts.recursive && opts.preserveRoot && isRootPath(path) {
			_, _ = fmt.Fprintf(stdio.Err, "rm: it is dangerous to operate recursively on '/'\n")
			_, _ = fmt.Fprintf(stdio.Err, "rm: use --no-preserve-root to override this failsafe\n")
			return command.SilentFailure()
		}
		if !confirm(stdio, in, path, opts) {
			return nil
		}
		if opts.recursive {
			if opts.oneFileSystem {
				dev, derr := deviceOf(path)
				if derr != nil {
					return derr
				}
				if err := removeTree(path, dev); err != nil {
					return err
				}
			} else {
				if err := os.RemoveAll(path); err != nil {
					return err
				}
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
	_, _ = fmt.Fprintf(stdio.Err, "rm: remove '%s'? ", path)
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
		_, _ = fmt.Fprintf(stdio.Out, "removed '%s'\n", path)
	}
}

// isRootPath reports whether path resolves to the filesystem root "/". It
// cleans the operand (so ".", trailing slashes and "/../" are handled) and,
// when possible, also resolves it to an absolute path so that operands such as
// "/." or symlinks pointing at "/" are caught by --preserve-root.
func isRootPath(path string) bool {
	if filepath.Clean(path) == "/" {
		return true
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return filepath.Clean(abs) == "/"
}

// deviceOf returns the filesystem device id (st_dev) for path. It is used by
// --one-file-system to detect when recursion would cross a mount point.
func deviceOf(path string) (uint64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("cannot determine device of %s", path)
	}
	return uint64(st.Dev), nil
}

// removeTree recursively removes path, but with --one-file-system it skips any
// subdirectory that lives on a different filesystem than dev (the device of the
// top-level argument). Skipped directories are left in place, which means the
// parent directory cannot be removed either; that mirrors GNU rm's behaviour.
func removeTree(path string, dev uint64) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return os.Remove(path)
	}

	// A directory on a different device than the top argument is skipped.
	if d, derr := deviceOf(path); derr == nil && d != dev {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	skipped := false
	for _, entry := range entries {
		child := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if d, derr := deviceOf(child); derr == nil && d != dev {
				// Different filesystem: leave it (and thus its parent) in place.
				skipped = true
				continue
			}
		}
		if err := removeTree(child, dev); err != nil {
			return err
		}
	}

	if skipped {
		// Cannot remove the directory while a foreign mount remains beneath it.
		return nil
	}
	return os.Remove(path)
}
