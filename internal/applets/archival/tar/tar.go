// Package tar implements the tar applet: create, list and extract POSIX tar
// archives, optionally gzip-compressed with -z. It covers the everyday
// "tar -czf", "tar -tzf" and "tar -xzf" workflows on top of the standard
// archive/tar and compress/gzip packages.
package tar

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tar applet.
type Command struct{}

// New returns a tar command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tar" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Archive files (create, list, extract)" }

type options struct {
	create  bool
	extract bool
	list    bool
	gzip    bool
	verbose bool
	file    string
	dir     string
}

// Run executes tar.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-c|-x|-t] [-z] [-f ARCHIVE] [FILE]...", stdio.Err)
	create := fs.BoolP("create", "c", false, "create a new archive")
	extract := fs.BoolP("extract", "x", false, "extract files from an archive")
	list := fs.BoolP("list", "t", false, "list the contents of an archive")
	gz := fs.BoolP("gzip", "z", false, "filter the archive through gzip")
	verbose := fs.BoolP("verbose", "v", false, "verbosely list files processed")
	file := fs.StringP("file", "f", "", "use archive file (default: stdin/stdout)")
	dir := fs.StringP("directory", "C", "", "change to directory before extracting/creating")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		create: *create, extract: *extract, list: *list,
		gzip: *gz, verbose: *verbose, file: *file, dir: *dir,
	}

	mode := 0
	for _, b := range []bool{opts.create, opts.extract, opts.list} {
		if b {
			mode++
		}
	}
	if mode != 1 {
		_, _ = fmt.Fprintln(stdio.Err, "tar: you must specify exactly one of -c, -x, -t")
		return command.SilentFailure()
	}

	var runErr error
	switch {
	case opts.create:
		runErr = c.doCreate(stdio, opts, fs.Args())
	case opts.list:
		runErr = c.doList(stdio, opts)
	case opts.extract:
		runErr = c.doExtract(stdio, opts)
	}
	if runErr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "tar: %v\n", runErr)
		return command.SilentFailure()
	}
	return nil
}

// doCreate writes an archive of the given paths to the -f file (or stdout).
func (c *Command) doCreate(stdio command.IO, opts options, paths []string) error {
	if len(paths) == 0 {
		return fmt.Errorf("refusing to create empty archive")
	}

	w := stdio.Out
	if opts.file != "" && opts.file != "-" {
		f, err := os.Create(opts.file) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		w = f
	}

	if opts.gzip {
		gw := gzip.NewWriter(w)
		defer func() { _ = gw.Close() }()
		w = gw
	}

	tw := tar.NewWriter(w)
	defer func() { _ = tw.Close() }()

	base := opts.dir
	for _, p := range paths {
		if err := addPath(stdio, tw, base, p, opts.verbose); err != nil {
			return err
		}
	}
	return tw.Close()
}

// addPath walks p (relative to base) and writes every entry to tw.
func addPath(stdio command.IO, tw *tar.Writer, base, p string, verbose bool) error {
	root := p
	if base != "" {
		root = filepath.Join(base, p)
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := path
		if base != "" {
			r, rerr := filepath.Rel(base, path)
			if rerr != nil {
				return rerr
			}
			rel = r
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if verbose {
			_, _ = fmt.Fprintln(stdio.Err, hdr.Name)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(path) //nolint:gosec // archiving a user-named file
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.Copy(tw, f) //nolint:gosec // copying user file into archive
		return err
	})
}

// reader opens the archive for reading: the -f file or stdin, plus a gzip layer
// when -z is set. It returns the reader and a close function.
func (c *Command) reader(stdio command.IO, opts options) (*tar.Reader, func() error, error) {
	r := stdio.In
	closers := []func() error{}
	if opts.file != "" && opts.file != "-" {
		f, err := os.Open(opts.file) //nolint:gosec // user-named file
		if err != nil {
			return nil, nil, err
		}
		closers = append(closers, f.Close)
		r = f
	}
	if opts.gzip {
		gr, err := gzip.NewReader(r)
		if err != nil {
			for _, cl := range closers {
				_ = cl()
			}
			return nil, nil, err
		}
		closers = append(closers, gr.Close)
		r = gr
	}
	closeAll := func() error {
		for i := len(closers) - 1; i >= 0; i-- {
			_ = closers[i]()
		}
		return nil
	}
	return tar.NewReader(r), closeAll, nil
}

// doList prints the name of every entry in the archive.
func (c *Command) doList(stdio command.IO, opts options) error {
	tr, closeAll, err := c.reader(stdio, opts)
	if err != nil {
		return err
	}
	defer func() { _ = closeAll() }()

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(stdio.Out, hdr.Name)
	}
}

// doExtract writes every entry in the archive to disk under the -C directory
// (or the current directory).
func (c *Command) doExtract(stdio command.IO, opts options) error {
	tr, closeAll, err := c.reader(stdio, opts)
	if err != nil {
		return err
	}
	defer func() { _ = closeAll() }()

	dest := opts.dir
	if dest == "" {
		dest = "."
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := extractEntry(stdio, tr, dest, hdr, opts.verbose); err != nil {
			return err
		}
	}
}

// extractEntry writes a single archive entry safely under dest, rejecting paths
// that would escape the destination directory.
func extractEntry(stdio command.IO, tr *tar.Reader, dest string, hdr *tar.Header, verbose bool) error {
	target, err := safeJoin(dest, hdr.Name)
	if err != nil {
		return err
	}
	if verbose {
		_, _ = fmt.Fprintln(stdio.Err, hdr.Name)
	}

	switch hdr.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, fs.FileMode(hdr.Mode)) //nolint:gosec // archive-defined mode
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.FileMode(hdr.Mode)) //nolint:gosec // archive-defined mode
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil { //nolint:gosec // extracting user archive
			_ = f.Close()
			return err
		}
		return f.Close()
	case tar.TypeSymlink:
		return os.Symlink(hdr.Linkname, target)
	default:
		// Skip unsupported entry types (devices, fifos) silently.
		return nil
	}
}

// safeJoin joins dest and name, ensuring the result stays within dest to defend
// against path-traversal entries ("../../etc/passwd").
func safeJoin(dest, name string) (string, error) {
	target := filepath.Join(dest, name)
	cleanDest := filepath.Clean(dest) + string(os.PathSeparator)
	if target != filepath.Clean(dest) && !hasPrefix(target, cleanDest) {
		return "", fmt.Errorf("entry %q is outside the destination directory", name)
	}
	return target, nil
}

// hasPrefix reports whether s starts with prefix.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
