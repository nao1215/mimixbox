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
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
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

// Run executes bzip2 by delegating the shared file-handling model to the comp
// frontend; this Command only supplies the codec, the -t test mode and the
// naming rules.
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

	cfg := comp.Config{
		Name:                c.Name(),
		Transform:           transform,
		Test:                testStream,
		OutputName:          outputName,
		RemoveOutputOnError: true,
	}
	opts := comp.Options{
		Decompress: *decompress || *test,
		Stdout:     *stdout,
		Keep:       *keep,
		Force:      *force,
		Test:       *test,
	}
	return cfg.Run(stdio, opts, fs.Args())
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
