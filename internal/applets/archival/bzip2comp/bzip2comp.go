// Package bzip2comp implements the bzip2 applet: a full bzip2 compressor and
// decompressor that mirrors the upstream bzip2 command-line interface. Unlike
// the decompress-only bunzip2 applet (which relies on the standard library's
// compress/bzip2), this applet uses github.com/dsnet/compress/bzip2 so it can
// both compress and decompress, giving a true round trip. It handles
// stdin/stdout and file operands with the usual -c, -d, -k, -f and -t options.
package bzip2comp

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/nao1215/mimixbox/internal/command"
)

// suffix is the filename extension bzip2 appends to compressed files.
const suffix = ".bz2"

// Command is the bzip2 applet.
type Command struct{}

// New returns a bzip2 command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "bzip2" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compress or decompress files (.bz2)" }

type options struct {
	decompress bool
	stdout     bool
	keep       bool
	force      bool
	test       bool
}

// Run executes bzip2.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(c.help())
	decompress := fs.BoolP("decompress", "d", false, "decompress")
	stdout := fs.BoolP("stdout", "c", false, "write to standard output and keep the input files")
	keep := fs.BoolP("keep", "k", false, "keep (don't delete) input files")
	force := fs.BoolP("force", "f", false, "force overwrite of the output file")
	test := fs.BoolP("test", "t", false, "test integrity of compressed files")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		decompress: *decompress || *test,
		stdout:     *stdout,
		keep:       *keep,
		force:      *force,
		test:       *test,
	}

	files := fs.Args()
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		return c.runStream(stdio, opts)
	}

	var failed bool
	for _, f := range files {
		if err := c.processFile(stdio, f, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// runStream handles the stdin/stdout (no FILE) case.
func (c *Command) runStream(stdio command.IO, opts options) error {
	if opts.test {
		if err := testStream(stdio.In); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
		return nil
	}
	if err := transform(stdio.In, stdio.Out, opts.decompress); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}
	return nil
}

// processFile compresses, decompresses or tests one file.
func (c *Command) processFile(stdio command.IO, name string, opts options) error {
	if opts.test {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		if err := testStream(in); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}

	if opts.stdout {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		return transform(in, stdio.Out, opts.decompress)
	}

	out, err := outputName(name, opts.decompress)
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
	if err := transform(in, w, opts.decompress); err != nil {
		_ = w.Close()
		_ = os.Remove(out)
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

// outputName derives the output filename: add .bz2 when compressing, strip it
// when decompressing.
func outputName(name string, decompress bool) (string, error) {
	if decompress {
		switch {
		case strings.HasSuffix(name, suffix):
			return strings.TrimSuffix(name, suffix), nil
		case strings.HasSuffix(name, ".tbz2"):
			return strings.TrimSuffix(name, ".tbz2") + ".tar", nil
		case strings.HasSuffix(name, ".tbz"):
			return strings.TrimSuffix(name, ".tbz") + ".tar", nil
		default:
			return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
		}
	}
	return name + suffix, nil
}

// transform copies r to w, bzip2-compressing or -decompressing along the way.
func transform(r io.Reader, w io.Writer, decompress bool) error {
	if decompress {
		zr, err := bzip2.NewReader(r, nil)
		if err != nil {
			return err
		}
		defer func() { _ = zr.Close() }()
		if _, err := io.Copy(w, zr); err != nil { //nolint:gosec // decompressing user data
			return err
		}
		return nil
	}
	zw, err := bzip2.NewWriter(w, nil)
	if err != nil {
		return err
	}
	if _, err := io.Copy(zw, r); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

// testStream verifies that r is a valid bzip2 stream by decompressing it and
// discarding the output.
func testStream(r io.Reader) error {
	zr, err := bzip2.NewReader(r, nil)
	if err != nil {
		return err
	}
	defer func() { _ = zr.Close() }()
	if _, err := io.Copy(io.Discard, zr); err != nil { //nolint:gosec // verifying user data
		return err
	}
	return nil
}

func (c *Command) help() command.Help {
	return command.Help{
		Description: "Compress each FILE in place to FILE.bz2 using the bzip2 algorithm; with -d (or -t) decompress instead. With -c write to standard output and keep the input. When no FILE is given (or FILE is '-') read standard input and write standard output.",
		Examples: []command.Example{
			{Command: "bzip2 file", Explain: "Compress 'file' to 'file.bz2', removing 'file'."},
			{Command: "bzip2 -k file", Explain: "Compress and keep the original 'file'."},
			{Command: "bzip2 -dc file.bz2", Explain: "Decompress to standard output."},
			{Command: "bzip2 -t file.bz2", Explain: "Test the integrity of 'file.bz2'."},
		},
		ExitStatus: "0  success.\n1  a file could not be read, written, or decoded.",
	}
}
