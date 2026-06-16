// Package bunzip2 implements the bunzip2 applet: decompress bzip2 (.bz2) files,
// or standard input when no file is given. Go's standard library provides a
// bzip2 decompressor (but no compressor), so this is decompress-only.
package bunzip2

import (
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the bunzip2 applet.
type Command struct{}

// New returns a bunzip2 command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "bunzip2" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Decompress bzip2 (.bz2) files" }

// Run executes bunzip2 by delegating the shared file-handling model to the comp
// frontend in decompress-only mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Decompress each bzip2 (.bz2) FILE, or standard input when no FILE (or\n" +
			"\"-\") is given. A decompressed file replaces its .bz2 original unless -k is\n" +
			"given to keep it, or -c is given to write to standard output instead.",
		Examples: []command.Example{
			{Command: "bunzip2 archive.tar.bz2", Explain: "decompress to archive.tar and remove the .bz2"},
			{Command: "bunzip2 -k data.bz2", Explain: "decompress but keep the original"},
			{Command: "bunzip2 -c data.bz2", Explain: "write the decompressed data to standard output"},
		},
		ExitStatus: "0  success.\n1  a file lacked a known suffix, could not be read, or the output already exists without -f.",
	})
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
	}
	opts := comp.Options{Decompress: true, Keep: *keep, Stdout: *stdout, Force: *force}
	return cfg.Run(stdio, opts, fs.Args())
}

// transform decompresses r into w. bunzip2 only ever decompresses, so the
// decompress flag from the shared frontend is always true and ignored here.
func transform(r io.Reader, w io.Writer, _ bool) error {
	return decompressStream(r, w)
}

// outputName strips a .bz2/.tbz2 suffix to derive the decompressed file name.
// bunzip2 only decompresses, so the decompress flag is always true.
func outputName(name string, _ bool) (string, error) {
	switch {
	case strings.HasSuffix(name, ".bz2"):
		return strings.TrimSuffix(name, ".bz2"), nil
	case strings.HasSuffix(name, ".tbz2"):
		return strings.TrimSuffix(name, ".tbz2") + ".tar", nil
	case strings.HasSuffix(name, ".tbz"):
		return strings.TrimSuffix(name, ".tbz") + ".tar", nil
	default:
		return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
	}
}

// decompressStream copies r, bzip2-decompressed, into w.
func decompressStream(r io.Reader, w io.Writer) error {
	if _, err := io.Copy(w, bzip2.NewReader(r)); err != nil { //nolint:gosec // decompressing user data
		return err
	}
	return nil
}
