// Package tree implements the tree applet: list the contents of directories in a
// tree-like format.
package tree

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tree applet.
type Command struct{}

// New returns a tree command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tree" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List directory contents in a tree-like format" }

type options struct {
	all      bool
	dirsOnly bool
	maxLevel int
}

// counts accumulates the directory and file totals for the summary line.
type counts struct {
	dirs  int
	files int
}

// Run executes tree.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [DIRECTORY]...", stdio.Err).WithHelp(command.Help{
		Description: "Print each DIRECTORY (default the current directory) as an indented tree, " +
			"followed by a count of the directories and files shown.",
		Examples: []command.Example{
			{Command: "tree", Explain: "Show the current directory tree."},
			{Command: "tree -L 2 -d /etc", Explain: "Show directories only, two levels deep."},
		},
		ExitStatus: "0  success.\n1  a directory could not be read.",
	})
	all := fs.BoolP("all", "a", false, "list entries starting with a dot")
	dirsOnly := fs.BoolP("dirs-only", "d", false, "list directories only")
	level := fs.IntP("level", "L", -1, "descend only LEVEL directories deep")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	roots := fs.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	opts := options{all: *all, dirsOnly: *dirsOnly, maxLevel: *level}

	var total counts
	var failed bool
	for _, root := range roots {
		_, _ = fmt.Fprintln(stdio.Out, root)
		total.dirs++ // the root directory itself is counted
		if err := walk(stdio.Out, root, "", 1, opts, &total); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tree: %v\n", err)
			failed = true
		}
	}

	if opts.dirsOnly {
		_, _ = fmt.Fprintf(stdio.Out, "\n%s\n", plural(total.dirs, "directory", "directories"))
	} else {
		_, _ = fmt.Fprintf(stdio.Out, "\n%s, %s\n", plural(total.dirs, "directory", "directories"), plural(total.files, "file", "files"))
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// plural formats a count with the singular or plural noun.
func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

// walk prints the children of dir, prefixed for the tree connectors.
func walk(w io.Writer, dir, prefix string, depth int, opts options, total *counts) error {
	if opts.maxLevel >= 0 && depth > opts.maxLevel {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	filtered := entries[:0]
	for _, e := range entries {
		if !opts.all && len(e.Name()) > 0 && e.Name()[0] == '.' {
			continue
		}
		if opts.dirsOnly && !e.IsDir() {
			continue
		}
		filtered = append(filtered, e)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Name() < filtered[j].Name() })

	for i, e := range filtered {
		last := i == len(filtered)-1
		connector, childPrefix := "├── ", prefix+"│   "
		if last {
			connector, childPrefix = "└── ", prefix+"    "
		}
		_, _ = fmt.Fprintf(w, "%s%s%s\n", prefix, connector, e.Name())
		if e.IsDir() {
			total.dirs++
			if err := walk(w, filepath.Join(dir, e.Name()), childPrefix, depth+1, opts, total); err != nil {
				return err
			}
		} else {
			total.files++
		}
	}
	return nil
}
