// Package gunzip implements the gunzip applet: decompress gzip (.gz) files, or
// standard input when no file is given. It is the decompress-only companion to
// gzip.
package gunzip

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the gunzip applet.
type Command struct{}

// New returns a gunzip command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "gunzip" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Decompress gzip (.gz) files" }

type options struct {
	keep   bool
	stdout bool
	force  bool
}

// Run executes gunzip.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Decompress each gzip (.gz) FILE, replacing FILE.gz with FILE. With no FILE, or with -, read " +
			"standard input and write the decompressed data to standard output.",
		Examples: []command.Example{
			{Command: "gunzip archive.gz", Explain: "Decompress archive.gz to archive and remove the .gz file."},
			{Command: "gunzip -c archive.gz", Explain: "Write the decompressed data to standard output."},
			{Command: "gunzip -k archive.gz", Explain: "Decompress but keep the original .gz file."},
		},
		ExitStatus: "0  all files were decompressed successfully.\n1  a file was missing, had a bad stream, or could not be written.",
	})
	keep := fs.BoolP("keep", "k", false, "keep (don't delete) input files")
	stdout := fs.BoolP("stdout", "c", false, "write on standard output, keep original files unchanged")
	force := fs.BoolP("force", "f", false, "force overwrite of output file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	opts := options{keep: *keep, stdout: *stdout, force: *force}

	files := fs.Args()
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		if err := decompressStream(stdio.In, stdio.Out); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "gunzip: %v\n", err)
			return command.SilentFailure()
		}
		return nil
	}

	var failed bool
	for _, f := range files {
		if err := c.processFile(stdio, f, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "gunzip: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// processFile decompresses one .gz file. With -c the result goes to stdout;
// otherwise FILE.gz becomes FILE and (unless -k) the input is removed.
func (c *Command) processFile(stdio command.IO, name string, opts options) error {
	if opts.stdout {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		return decompressStream(in, stdio.Out)
	}

	out, err := outputName(name)
	if err != nil {
		return err
	}
	if !opts.force {
		if _, statErr := os.Stat(out); statErr == nil {
			return fmt.Errorf("%s already exists; use -f to overwrite", out)
		}
	}

	in, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	w, err := os.Create(out) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	if err := decompressStream(in, w); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if !opts.keep {
		return os.Remove(name)
	}
	return nil
}

// outputName strips a .gz/.tgz suffix to derive the decompressed file name.
func outputName(name string) (string, error) {
	switch {
	case strings.HasSuffix(name, ".gz"):
		return strings.TrimSuffix(name, ".gz"), nil
	case strings.HasSuffix(name, ".tgz"):
		return strings.TrimSuffix(name, ".tgz") + ".tar", nil
	default:
		return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
	}
}

// decompressStream copies r, gzip-decompressed, into w.
func decompressStream(r io.Reader, w io.Writer) error {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	if _, err := io.Copy(w, zr); err != nil { //nolint:gosec // decompressing user data
		return err
	}
	return nil
}
