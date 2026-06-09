// Package man implements the man applet: locate and display a manual page from
// the manual path. It finds plain or gzip-compressed pages and reports clearly
// when a page is missing.
package man

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the man applet.
type Command struct{}

// New returns a man command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "man" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display a manual page" }

// notFoundCode matches util-linux/man-db's exit status for a missing page.
const notFoundCode = 16

// Run executes man.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-M PATH] [SECTION] PAGE", stdio.Err).WithHelp(command.Help{
		Description: "Locate and print a manual PAGE. The optional SECTION (a leading number) limits " +
			"the search. Pages are looked up under -M PATH, then $MANPATH, then the standard " +
			"system manual directories; both plain and .gz pages are supported.",
		Examples: []command.Example{
			{Command: "man ls", Explain: "Show the ls page from any section."},
			{Command: "man 5 passwd", Explain: "Show the section 5 passwd page."},
			{Command: "man -M ./man tool", Explain: "Search ./man for the tool page."},
		},
		ExitStatus: "0  the page was found.\n16  no manual entry was found.",
		Notes: []string{
			"The page is shown as stored; roff/groff formatting is not applied.",
		},
	})
	manpath := fs.StringP("manpath", "M", "", "search PATH for manual pages")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	section, page := "", ""
	switch len(operands) {
	case 0:
		_, _ = fmt.Fprintln(stdio.Err, "What manual page do you want?")
		return command.SilentFailure()
	case 1:
		page = operands[0]
	default:
		if isSection(operands[0]) {
			section, page = operands[0], operands[1]
		} else {
			page = operands[0]
		}
	}

	path, err := find(searchPaths(*manpath), section, page)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "No manual entry for %s\n", page)
		return &command.ExitError{Code: notFoundCode}
	}

	if err := display(stdio.Out, path); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "man: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// isSection reports whether s looks like a manual section (1, 3, 3p, ...).
func isSection(s string) bool {
	if s == "" || s[0] < '0' || s[0] > '9' {
		return false
	}
	return true
}

// searchPaths returns the manual directories to search, in priority order.
func searchPaths(override string) []string {
	if override != "" {
		return splitPath(override)
	}
	if env := os.Getenv("MANPATH"); env != "" {
		return splitPath(env)
	}
	return []string{"/usr/share/man", "/usr/local/share/man", "/usr/local/man"}
}

func splitPath(p string) []string {
	var out []string
	for _, part := range strings.Split(p, ":") {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// find locates the page file under one of roots. When section is empty every
// section directory is tried in order.
func find(roots []string, section, page string) (string, error) {
	sections := []string{section}
	if section == "" {
		sections = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "3p", "1p"}
	}
	for _, root := range roots {
		for _, sec := range sections {
			dir := filepath.Join(root, "man"+sectionDir(sec))
			for _, cand := range []string{
				filepath.Join(dir, page+"."+sec),
				filepath.Join(dir, page+"."+sec+".gz"),
			} {
				if _, err := os.Stat(cand); err == nil {
					return cand, nil
				}
			}
		}
	}
	return "", os.ErrNotExist
}

// sectionDir maps a section like "3p" to its directory suffix "3".
func sectionDir(sec string) string {
	if sec == "" {
		return ""
	}
	return sec[:1]
}

// display writes the page to out, decompressing a .gz page on the way.
func display(out io.Writer, path string) error {
	f, err := os.Open(path) //nolint:gosec // resolved manual page path
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var r io.Reader = bufio.NewReader(f)
	if strings.HasSuffix(path, ".gz") {
		zr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer func() { _ = zr.Close() }()
		r = zr
	}
	_, err = io.Copy(out, r) //nolint:gosec // copying a local manual page
	return err
}
