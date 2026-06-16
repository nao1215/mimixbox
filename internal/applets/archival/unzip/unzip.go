// Package unzip implements the unzip applet: list or extract the contents of a
// ZIP archive using the standard archive/zip package. It covers "unzip a.zip",
// "unzip -l a.zip" and "unzip -d DIR a.zip".
package unzip

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the unzip applet.
type Command struct{}

// New returns an unzip command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unzip" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Extract files from a ZIP archive" }

type options struct {
	list    bool
	dir     string
	verbose bool
}

// Run executes unzip.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... ARCHIVE", stdio.Err).WithHelp(command.Help{
		Description: "List or extract the contents of a ZIP ARCHIVE. By default every entry is extracted into the " +
			"current directory; -l lists entries instead, and -d chooses a destination directory.",
		Examples: []command.Example{
			{Command: "unzip files.zip", Explain: "Extract every entry into the current directory."},
			{Command: "unzip -l files.zip", Explain: "List the archive contents without extracting."},
			{Command: "unzip -d out files.zip", Explain: "Extract the archive into the out directory."},
		},
		ExitStatus: "0  the archive was listed or extracted successfully.\n1  the archive was missing, malformed, or could not be extracted.",
	})
	list := fs.BoolP("list", "l", false, "list archive contents without extracting")
	dir := fs.StringP("dir", "d", "", "extract files into this directory")
	verbose := fs.BoolP("verbose", "v", false, "print each file as it is extracted")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	opts := options{list: *list, dir: *dir, verbose: *verbose}

	operands := fs.Args()
	if len(operands) != 1 {
		_, _ = fmt.Fprintln(stdio.Err, "unzip: usage: unzip [OPTION]... ARCHIVE")
		return command.SilentFailure()
	}

	if err := c.process(stdio, operands[0], opts); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "unzip: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// process opens the archive and either lists or extracts it.
func (c *Command) process(stdio command.IO, archive string, opts options) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()

	dest := opts.dir
	if dest == "" {
		dest = "."
	}

	for _, f := range zr.File {
		if opts.list {
			_, _ = fmt.Fprintf(stdio.Out, "%9d  %s\n", f.UncompressedSize64, f.Name)
			continue
		}
		if err := extractFile(stdio, f, dest, opts.verbose); err != nil {
			return err
		}
	}
	return nil
}

// extractFile writes one archive entry to disk under dest, rejecting paths that
// would escape the destination directory.
func extractFile(stdio command.IO, f *zip.File, dest string, verbose bool) error {
	target, err := safeJoin(dest, f.Name)
	if err != nil {
		return err
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, f.Mode())
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode()) //nolint:gosec // archive-defined mode
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, rc); err != nil { //nolint:gosec // extracting user archive
		_ = out.Close()
		return err
	}
	if verbose {
		_, _ = fmt.Fprintf(stdio.Err, " extracting: %s\n", f.Name)
	}
	return out.Close()
}

// safeJoin joins dest and name, ensuring the result stays within dest.
func safeJoin(dest, name string) (string, error) {
	target := filepath.Join(dest, name)
	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
	if target != filepath.Clean(dest) && len(target) < len(cleanDest) {
		return "", fmt.Errorf("entry %q is outside the destination directory", name)
	}
	if target != filepath.Clean(dest) && target[:len(cleanDest)] != cleanDest {
		return "", fmt.Errorf("entry %q is outside the destination directory", name)
	}
	return target, nil
}
