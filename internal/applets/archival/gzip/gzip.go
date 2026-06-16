// Package gzipCmd implements the gzip applet: compress or uncompress files
// using the DEFLATE algorithm, with the common GNU options. By default each
// FILE is compressed in place (replaced by FILE.gz); with -d the FILE.gz is
// decompressed back to FILE.
package gzipCmd

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
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

// Run executes gzip by delegating the shared file-handling model to the comp
// frontend; this Command only supplies the DEFLATE codec and the naming, error
// and overwrite rules that are specific to gzip.
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

	cfg := comp.Config{
		Name:       c.Name(),
		Transform:  transform,
		OutputName: outputName,
		// gzip's "already exists" wording and its GNU-style per-file error
		// formatting differ from the other compressors, so override both.
		ExistsErr:   func(out string) error { return fmt.Errorf("%s already exists; not overwritten", out) },
		WrapFileErr: func(name string, err error) error { return errors.New(command.FileError(name, err)) },
	}
	opts := comp.Options{Decompress: *decompress, Keep: *keep, Stdout: *stdout, Force: *force}
	return cfg.Run(stdio, opts, fs.Args())
}

// transform copies r to w, gzip-compressing or -decompressing along the way.
func transform(r io.Reader, w io.Writer, decompress bool) error {
	if decompress {
		return decompressStream(r, w)
	}
	return compressStream(r, w)
}

// outputName derives the output filename: add .gz when compressing, strip it
// when decompressing (an unknown suffix is an error).
func outputName(name string, decompress bool) (string, error) {
	if decompress {
		dst := strings.TrimSuffix(name, ".gz")
		if dst == name {
			return "", fmt.Errorf("unknown suffix -- ignored")
		}
		return dst, nil
	}
	return name + ".gz", nil
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
