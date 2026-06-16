// Package uncompress implements the uncompress applet: decompress .Z files
// produced by the classic Unix compress (or gzip). Each FILE.Z becomes FILE;
// with -c the result is written to standard output.
package uncompress

import (
	"context"
	"fmt"
	"os"
	"strings"

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

type options struct {
	stdout bool
	keep   bool
	force  bool
}

// Run executes uncompress.
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
	opts := options{stdout: *stdout, keep: *keep, force: *force}

	files := fs.Args()
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		if err := lzw.Decompress(stdio.In, stdio.Out); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "uncompress: %v\n", err)
			return command.SilentFailure()
		}
		return nil
	}

	var failed bool
	for _, f := range files {
		if err := c.processFile(stdio, f, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "uncompress: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// processFile decompresses one .Z file to FILE (or to stdout with -c).
func (c *Command) processFile(stdio command.IO, name string, opts options) error {
	in, err := os.Open(name) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if opts.stdout {
		return lzw.Decompress(in, stdio.Out)
	}

	if !strings.HasSuffix(name, ".Z") {
		return fmt.Errorf("%s: unknown suffix -- ignored", name)
	}
	out := strings.TrimSuffix(name, ".Z")
	if !opts.force {
		if _, statErr := os.Stat(out); statErr == nil {
			return fmt.Errorf("%s already exists; use -f to overwrite", out)
		}
	}

	w, err := os.Create(out) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	if err := lzw.Decompress(in, w); err != nil {
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
