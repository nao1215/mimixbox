// Package compress implements the compress applet: compress files with the
// classic Unix LZW algorithm, producing .Z files that the system
// compress/uncompress and gzip can read. Each FILE becomes FILE.Z; with -c the
// result is written to standard output and the input is left in place.
package compress

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
	"github.com/nao1215/mimixbox/internal/applets/archival/lzw"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the compress applet.
type Command struct{}

// New returns a compress command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "compress" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Compress files with LZW (.Z)" }

// Run executes compress by delegating the shared file-handling model to the
// comp frontend in compress-only mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Compress each FILE in place with the classic Unix LZW algorithm, replacing it with FILE.Z. " +
			"With -c write to standard output and keep the input. With no FILE (or FILE '-') read standard input.",
		Examples: []command.Example{
			{Command: "compress file.txt", Explain: "Replace file.txt with the compressed file.txt.Z."},
			{Command: "compress -k file.txt", Explain: "Create file.txt.Z but keep the original file.txt."},
			{Command: "compress -c file.txt > file.Z", Explain: "Write the compressed data to standard output."},
		},
		ExitStatus: "0  all files were compressed successfully.\n1  one or more files could not be compressed.",
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
	opts := comp.Options{Stdout: *stdout, Keep: *keep, Force: *force}
	return cfg.Run(stdio, opts, fs.Args())
}

// transform LZW-compresses r into w. compress only ever compresses, so the
// decompress flag from the shared frontend is always false and ignored here.
func transform(r io.Reader, w io.Writer, _ bool) error {
	return lzw.Compress(r, w)
}

// outputName appends the .Z suffix, rejecting a name that already ends in .Z.
// compress only compresses, so the decompress flag is always false.
func outputName(name string, _ bool) (string, error) {
	if strings.HasSuffix(name, ".Z") {
		return "", fmt.Errorf("%s already has .Z suffix -- unchanged", name)
	}
	return name + ".Z", nil
}
