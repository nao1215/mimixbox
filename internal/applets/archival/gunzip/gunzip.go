// Package gunzip implements the gunzip applet: decompress gzip (.gz) files, or
// standard input when no file is given. It is the decompress-only companion to
// gzip.
package gunzip

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
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

// Run executes gunzip by delegating the shared file-handling model to the comp
// frontend in decompress-only mode.
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

	cfg := comp.Config{
		Name:                c.Name(),
		Transform:           transform,
		OutputName:          outputName,
		RemoveOutputOnError: true, // don't leave a partial/empty output file behind on failure
	}
	opts := comp.Options{Decompress: true, Keep: *keep, Stdout: *stdout, Force: *force}
	return cfg.Run(stdio, opts, fs.Args())
}

// transform decompresses r into w. gunzip only ever decompresses, so the
// decompress flag from the shared frontend is always true and ignored here.
func transform(r io.Reader, w io.Writer, _ bool) error {
	return decompressStream(r, w)
}

// outputName strips a .gz/.tgz suffix to derive the decompressed file name.
// gunzip only decompresses, so the decompress flag is always true.
func outputName(name string, _ bool) (string, error) {
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
