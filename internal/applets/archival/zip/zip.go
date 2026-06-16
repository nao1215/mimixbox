// Package zip implements the zip applet: create a ZIP archive from files and
// directories. It is a focused subset built on the standard archive/zip
// package: "zip -r archive.zip dir file ...".
package zip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the zip applet.
type Command struct{}

// New returns a zip command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "zip" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Package and compress files into a ZIP archive" }

type options struct {
	recurse bool
	verbose bool
}

// Run executes zip.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... ARCHIVE FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Create the ZIP archive ARCHIVE containing the given FILE operands. Directories are " +
			"skipped unless -r is given, in which case they are added recursively.",
		Examples: []command.Example{
			{Command: "zip archive.zip file1 file2", Explain: "Create archive.zip from two files."},
			{Command: "zip -r archive.zip dir", Explain: "Recursively add a directory to the archive."},
			{Command: "zip -v archive.zip file", Explain: "Create the archive, listing each entry as it is added."},
		},
		ExitStatus: "0  the archive was created.\n1  an input was missing or could not be read.",
	})
	recurse := fs.BoolP("recurse-paths", "r", false, "recurse into directories")
	verbose := fs.BoolP("verbose", "v", false, "print each entry as it is added")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	opts := options{recurse: *recurse, verbose: *verbose}

	operands := fs.Args()
	if len(operands) < 2 {
		_, _ = fmt.Fprintln(stdio.Err, "zip: usage: zip [OPTION]... ARCHIVE FILE...")
		return command.SilentFailure()
	}
	archive, inputs := operands[0], operands[1:]

	if err := c.create(stdio, archive, inputs, opts); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "zip: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// create writes a ZIP archive containing the inputs.
func (c *Command) create(stdio command.IO, archive string, inputs []string, opts options) error {
	out, err := os.Create(archive) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	zw := zip.NewWriter(out)
	defer func() { _ = zw.Close() }()

	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if !opts.recurse {
				_, _ = fmt.Fprintf(stdio.Err, "zip: %s is a directory (use -r to recurse) -- skipped\n", in)
				continue
			}
			if err := addDir(stdio, zw, in, opts.verbose); err != nil {
				return err
			}
			continue
		}
		if err := addFile(stdio, zw, in, in, opts.verbose); err != nil {
			return err
		}
	}
	return zw.Close()
}

// addDir recursively adds every entry under root to the archive.
func addDir(stdio command.IO, zw *zip.Writer, root string, verbose bool) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return addFile(stdio, zw, path, path, verbose)
	})
}

// addFile copies one file into the archive under the name entry.
func addFile(stdio command.IO, zw *zip.Writer, path, entry string, verbose bool) error {
	f, err := os.Open(path) //nolint:gosec // archiving a user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w, err := zw.Create(filepath.ToSlash(entry))
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, f); err != nil { //nolint:gosec // copying user file
		return err
	}
	if verbose {
		_, _ = fmt.Fprintf(stdio.Err, "  adding: %s\n", entry)
	}
	return nil
}
