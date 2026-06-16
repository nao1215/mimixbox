// Package uncompress implements the uncompress applet: decompress .Z files
// produced by the classic Unix compress (or gzip). Each FILE.Z becomes FILE;
// with -c the result is written to standard output.
package uncompress

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
	"github.com/nao1215/mimixbox/internal/applets/archival/lzw"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the uncompress applet.
type Command struct{}

// New returns an uncompress command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uncompress" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Decompress LZW (.Z) files" }

// Run executes uncompress by delegating the shared file-handling model to the
// comp frontend in decompress-only mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Decompress each .Z FILE produced by the classic Unix compress (LZW), replacing FILE.Z with FILE. " +
			"With no FILE, or with -, read standard input and write to standard output.",
		Examples: []command.Example{
			{Command: "uncompress archive.Z", Explain: "Decompress archive.Z to archive and remove the .Z file."},
			{Command: "uncompress -c archive.Z", Explain: "Write the decompressed data to standard output."},
			{Command: "uncompress -k data.Z", Explain: "Decompress but keep the original .Z file."},
		},
		ExitStatus: "0  all files were decompressed successfully.\n1  a file was missing, had a bad stream, or could not be written.",
	})
	stdout := fs.BoolP("stdout", "c", false, "write on standard output, keep original files unchanged")
	keep := fs.BoolP("keep", "k", false, "keep (don't delete) input files")
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
	opts := comp.Options{Decompress: true, Stdout: *stdout, Keep: *keep, Force: *force}
	return cfg.Run(stdio, opts, fs.Args())
}

// transform LZW-decompresses r into w. uncompress only ever decompresses, so the
// decompress flag from the shared frontend is always true and ignored here.
func transform(r io.Reader, w io.Writer, _ bool) error {
	return lzw.Decompress(r, w)
}

// outputName strips the .Z suffix, rejecting a name that does not end in .Z.
// uncompress only decompresses, so the decompress flag is always true.
func outputName(name string, _ bool) (string, error) {
	if !strings.HasSuffix(name, ".Z") {
		return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
	}
	return strings.TrimSuffix(name, ".Z"), nil
}
