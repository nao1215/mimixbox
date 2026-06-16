// Package bunzip2 implements the bunzip2 applet: decompress bzip2 (.bz2) files,
// or standard input when no file is given. Go's standard library provides a
// bzip2 decompressor (but no compressor), so this is decompress-only.
package bunzip2

import (
	"compress/bzip2"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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

type options struct {
	keep   bool
	stdout bool
	force  bool
}

// Run executes bunzip2.
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
	opts := options{keep: *keep, stdout: *stdout, force: *force}

	files := fs.Args()
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		if err := decompressStream(stdio.In, stdio.Out); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "bunzip2: %v\n", err)
			return command.SilentFailure()
		}
		return nil
	}

	var failed bool
	for _, f := range files {
		if err := c.processFile(stdio, f, opts); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "bunzip2: %v\n", err)
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// processFile decompresses one .bz2 file.
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

// outputName strips a .bz2/.tbz2 suffix to derive the decompressed file name.
func outputName(name string) (string, error) {
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
