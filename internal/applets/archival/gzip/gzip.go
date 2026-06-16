// Package gzipCmd implements the gzip applet: compress or uncompress files
// using the DEFLATE algorithm, with the common GNU options. By default each
// FILE is compressed in place (replaced by FILE.gz); with -d the FILE.gz is
// decompressed back to FILE.
package gzipCmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the gzip applet.
type Command struct{}

// New returns a gzip command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "gzip" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Compress or uncompress FILEs (by default, compress FILES in-place)"
}

type options struct {
	decompress bool
	keep       bool
	stdout     bool
	force      bool
}

// Run executes gzip.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Compress each FILE in place using DEFLATE, replacing FILE with FILE.gz. With -d, decompress " +
			"FILE.gz back to FILE. With no FILE, or with -, read standard input and write to standard output.",
		Examples: []command.Example{
			{Command: "gzip notes.txt", Explain: "Compress notes.txt to notes.txt.gz and remove the original."},
			{Command: "gzip -d notes.txt.gz", Explain: "Decompress notes.txt.gz back to notes.txt."},
			{Command: "gzip -c notes.txt", Explain: "Write the compressed data to standard output, keeping the original."},
		},
		ExitStatus: "0  all files were processed successfully.\n1  a file was missing, had a bad stream, or could not be written.",
	})
	decompress := fs.BoolP("decompress", "d", false, "decompress")
	keep := fs.BoolP("keep", "k", false, "keep (don't delete) input files")
	stdout := fs.BoolP("stdout", "c", false, "write on standard output, keep original files unchanged")
	force := fs.BoolP("force", "f", false, "force overwrite of output file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		decompress: *decompress,
		keep:       *keep,
		stdout:     *stdout,
		force:      *force,
	}

	files := fs.Args()
	// With no operands (or "-"), read standard input and write to standard
	// output regardless of the other file-oriented options.
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		if err := streamStdio(stdio, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "gzip: %v\n", err)
			return command.SilentFailure()
		}
		return nil
	}

	return processFiles(stdio, files, opts)
}

// streamStdio compresses or decompresses standard input to standard output.
func streamStdio(stdio command.IO, opts options) error {
	if opts.decompress {
		return decompressStream(stdio.In, stdio.Out)
	}
	return compressStream(stdio.In, stdio.Out)
}

// processFiles handles each named operand, reporting failures on stderr without
// stopping the remaining files. The returned error only sets the exit code; its
// message has already been printed.
func processFiles(stdio command.IO, files []string, opts options) error {
	var firstErr error
	for _, name := range files {
		if name == "-" {
			if err := streamStdio(stdio, opts); err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "gzip: %v\n", err)
				firstErr = keepErr(firstErr)
			}
			continue
		}
		if err := processFile(stdio, name, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "gzip: %s\n", command.FileError(name, err))
			firstErr = keepErr(firstErr)
		}
	}
	return firstErr
}

// processFile compresses or decompresses a single named file.
func processFile(stdio command.IO, name string, opts options) error {
	if opts.decompress {
		return decompressFile(stdio, name, opts)
	}
	return compressFile(stdio, name, opts)
}

// compressFile compresses name into name.gz. Without -c, the original is
// removed unless -k is given.
func compressFile(stdio command.IO, name string, opts options) error {
	src, err := os.Open(name) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	if opts.stdout {
		return compressStream(src, stdio.Out)
	}

	dst := name + ".gz"
	out, err := createOutput(dst, opts.force)
	if err != nil {
		return err
	}
	if err := compressStream(src, out); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = src.Close()
	if !opts.keep {
		return os.Remove(name)
	}
	return nil
}

// decompressFile decompresses name (expected to end in .gz) into name without
// the .gz suffix. Without -c, the input is removed unless -k is given.
func decompressFile(stdio command.IO, name string, opts options) error {
	src, err := os.Open(name) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	if opts.stdout {
		return decompressStream(src, stdio.Out)
	}

	dst := strings.TrimSuffix(name, ".gz")
	if dst == name {
		return fmt.Errorf("unknown suffix -- ignored")
	}
	out, err := createOutput(dst, opts.force)
	if err != nil {
		return err
	}
	if err := decompressStream(src, out); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = src.Close()
	if !opts.keep {
		return os.Remove(name)
	}
	return nil
}

// createOutput creates dst. Without -f it refuses to overwrite an existing
// file, matching GNU gzip when standard input is not a terminal.
func createOutput(dst string, force bool) (*os.File, error) {
	if !force {
		if _, err := os.Stat(dst); err == nil {
			return nil, fmt.Errorf("%s already exists; not overwritten", dst)
		}
	}
	return os.Create(dst) //nolint:gosec // operating on a user-named file is the whole point
}

// compressStream writes the gzip-compressed form of r to w.
func compressStream(r io.Reader, w io.Writer) error {
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := io.Copy(gw, r); err != nil {
		_ = gw.Close()
		return err
	}
	return gw.Close()
}

// decompressStream writes the gzip-decompressed form of r to w.
func decompressStream(r io.Reader, w io.Writer) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gr.Close() }()
	_, err = io.Copy(w, gr) //nolint:gosec // decompressing a user-supplied file is the whole point
	return err
}

func keepErr(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
