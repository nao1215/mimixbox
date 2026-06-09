// Package find implements the find applet: walk one or more directory trees and
// print the entries that satisfy the given tests (-name, -type, -maxdepth, ...).
// It covers the predicates people reach for most; the default action is -print.
package find

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/version"
)

// Command is the find applet.
type Command struct{}

// New returns a find command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "find" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Search for files in a directory hierarchy" }

// predicate is one parsed test or action from the expression.
type predicate struct {
	kind  string
	value string
}

// config is the parsed expression: the roots to walk and the predicates to
// apply to every entry.
type config struct {
	roots    []string
	preds    []predicate
	maxDepth int
	minDepth int
	print0   bool
}

// Run executes find.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	// find has its own non-getopt expression grammar, so handle --help/--version
	// by hand before splitting paths from the expression. --version prints the
	// version line (not usage), matching every other applet's contract.
	for _, a := range args {
		switch a {
		case "--help":
			printUsage(stdio.Out)
			return nil
		case "--version":
			version.Print(stdio.Out, c.Name())
			return nil
		}
	}

	cfg, err := parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "find: %v\n", err)
		return command.SilentFailure()
	}

	var failed bool
	for _, root := range cfg.roots {
		if walkErr := walk(stdio, root, cfg); walkErr != nil {
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// parse splits args into the leading path operands and the trailing expression.
func parse(args []string) (config, error) {
	cfg := config{maxDepth: -1, minDepth: 0}

	i := 0
	for i < len(args) && !strings.HasPrefix(args[i], "-") && args[i] != "!" {
		cfg.roots = append(cfg.roots, args[i])
		i++
	}
	if len(cfg.roots) == 0 {
		cfg.roots = []string{"."}
	}

	for i < len(args) {
		tok := args[i]
		switch tok {
		case "-name", "-iname", "-path", "-type":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing argument to '%s'", tok)
			}
			cfg.preds = append(cfg.preds, predicate{kind: tok, value: args[i+1]})
			i += 2
		case "-maxdepth", "-mindepth":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("missing argument to '%s'", tok)
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n < 0 {
				return cfg, fmt.Errorf("invalid argument '%s' to '%s'", args[i+1], tok)
			}
			if tok == "-maxdepth" {
				cfg.maxDepth = n
			} else {
				cfg.minDepth = n
			}
			i += 2
		case "-empty":
			cfg.preds = append(cfg.preds, predicate{kind: tok})
			i++
		case "-print":
			i++
		case "-print0":
			cfg.print0 = true
			i++
		default:
			return cfg, fmt.Errorf("unknown predicate '%s'", tok)
		}
	}
	return cfg, nil
}

// walk descends root, applying depth limits and predicates and printing every
// entry that matches.
func walk(stdio command.IO, root string, cfg config) error {
	var retErr error
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "find: %s\n", command.FileError(path, err))
			retErr = err
			return nil
		}

		depth := entryDepth(root, path)
		if cfg.maxDepth >= 0 && depth > cfg.maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if depth < cfg.minDepth {
			return nil
		}

		if matches(path, d, cfg.preds) {
			printPath(stdio, path, cfg.print0)
		}
		return nil
	})
	return retErr
}

// entryDepth reports how many path components deep path is relative to root,
// where root itself is depth 0.
func entryDepth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

// matches reports whether an entry satisfies every predicate (logical AND).
func matches(path string, d fs.DirEntry, preds []predicate) bool {
	for _, p := range preds {
		if !matchOne(path, d, p) {
			return false
		}
	}
	return true
}

// matchOne evaluates a single predicate against an entry.
func matchOne(path string, d fs.DirEntry, p predicate) bool {
	switch p.kind {
	case "-name":
		ok, _ := filepath.Match(p.value, filepath.Base(path))
		return ok
	case "-iname":
		ok, _ := filepath.Match(strings.ToLower(p.value), strings.ToLower(filepath.Base(path)))
		return ok
	case "-path":
		ok, _ := filepath.Match(p.value, path)
		return ok
	case "-type":
		return matchType(d, p.value)
	case "-empty":
		return isEmpty(path, d)
	}
	return false
}

// matchType reports whether the entry is of the requested type: f (regular),
// d (directory), l (symlink), p (FIFO), s (socket).
func matchType(d fs.DirEntry, t string) bool {
	m := d.Type()
	switch t {
	case "f":
		return m.IsRegular()
	case "d":
		return d.IsDir()
	case "l":
		return m&fs.ModeSymlink != 0
	case "p":
		return m&fs.ModeNamedPipe != 0
	case "s":
		return m&fs.ModeSocket != 0
	}
	return false
}

// isEmpty reports whether a regular file has zero size or a directory has no
// entries.
func isEmpty(path string, d fs.DirEntry) bool {
	if d.IsDir() {
		entries, err := readDirNames(path)
		return err == nil && len(entries) == 0
	}
	info, err := d.Info()
	return err == nil && info.Mode().IsRegular() && info.Size() == 0
}

// printPath writes a found path, terminated by NUL for -print0 or newline.
func printPath(stdio command.IO, path string, zero bool) {
	if zero {
		_, _ = fmt.Fprintf(stdio.Out, "%s\x00", path)
		return
	}
	_, _ = fmt.Fprintln(stdio.Out, path)
}

func printUsage(w interface{ Write([]byte) (int, error) }) {
	_, _ = w.Write([]byte("Usage: find [PATH]... [EXPRESSION]\n\n" +
		"Search each PATH (default: the current directory) recursively and act on the\n" +
		"entries that match EXPRESSION. With no expression, every entry is printed.\n\n" +
		"Tests: -name, -iname, -path, -type [f|d|l|p|s], -empty\n" +
		"Depth: -maxdepth N, -mindepth N\n" +
		"Actions: -print (default), -print0\n\n" +
		"Examples:\n" +
		"  find .                       Print every entry under the current directory.\n" +
		"  find . -name '*.go'          Find files whose name ends in .go.\n" +
		"  find /tmp -type d -empty     Find empty directories under /tmp.\n" +
		"  find . -type f -print0       Print file paths separated by NUL (for xargs -0).\n\n" +
		"Notes:\n" +
		"  - This is a subset of GNU find: the tests, depth limits, and actions listed above.\n" +
		"  - -print0 pairs with 'xargs -0' to handle paths that contain spaces or newlines.\n"))
}
