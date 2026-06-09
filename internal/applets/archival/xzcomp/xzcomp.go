// Package xzcomp implements the BusyBox compression applets that wrap a single
// stream codec: xz/unxz/xzcat and lzma/unlzma/lzcat (via github.com/ulikunitz/xz),
// plus the decompress-to-stdout aliases zcat (gzip) and bzcat (bzip2). They all
// share one file-handling model (stdin/stdout, -c, -k, -f, in-place suffix
// rename) so each applet is just a small configuration of that model.
package xzcomp

import (
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

// codec describes one compression format: its filename suffix and stream
// constructors. newWriter is nil for decompress-only formats (gzip/bzip2 here).
type codec struct {
	suffix    string
	newReader func(io.Reader) (io.ReadCloser, error)
	newWriter func(io.Writer) (io.WriteCloser, error)
}

func xzReader(r io.Reader) (io.ReadCloser, error) {
	zr, err := xz.NewReader(r)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(zr), nil
}

func xzWriter(w io.Writer) (io.WriteCloser, error) { return xz.NewWriter(w) }

func lzmaReader(r io.Reader) (io.ReadCloser, error) {
	zr, err := lzma.NewReader(r)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(zr), nil
}

func lzmaWriter(w io.Writer) (io.WriteCloser, error) { return lzma.NewWriter(w) }

func gzipReader(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) }

func bzip2Reader(r io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(bzip2.NewReader(r)), nil
}

var (
	xzCodec    = codec{suffix: ".xz", newReader: xzReader, newWriter: xzWriter}
	lzmaCodec  = codec{suffix: ".lzma", newReader: lzmaReader, newWriter: lzmaWriter}
	gzipCodec  = codec{suffix: ".gz", newReader: gzipReader}
	bzip2Codec = codec{suffix: ".bz2", newReader: bzip2Reader}
)

// Command is one compression applet.
type Command struct {
	name            string
	codec           codec
	forceDecompress bool // unxz/xzcat/zcat/... always decompress
	forceStdout     bool // *cat variants always write to stdout
}

// NewXz returns the xz applet (compresses by default).
func NewXz() *Command { return &Command{name: "xz", codec: xzCodec} }

// NewUnxz returns the unxz applet (xz -d).
func NewUnxz() *Command { return &Command{name: "unxz", codec: xzCodec, forceDecompress: true} }

// NewXzcat returns the xzcat applet (xz -dc).
func NewXzcat() *Command {
	return &Command{name: "xzcat", codec: xzCodec, forceDecompress: true, forceStdout: true}
}

// NewLzma returns the lzma applet (compresses by default).
func NewLzma() *Command { return &Command{name: "lzma", codec: lzmaCodec} }

// NewUnlzma returns the unlzma applet (lzma -d).
func NewUnlzma() *Command { return &Command{name: "unlzma", codec: lzmaCodec, forceDecompress: true} }

// NewLzcat returns the lzcat applet (lzma -dc).
func NewLzcat() *Command {
	return &Command{name: "lzcat", codec: lzmaCodec, forceDecompress: true, forceStdout: true}
}

// NewZcat returns the zcat applet (gunzip -c).
func NewZcat() *Command {
	return &Command{name: "zcat", codec: gzipCodec, forceDecompress: true, forceStdout: true}
}

// NewBzcat returns the bzcat applet (bunzip2 -c).
func NewBzcat() *Command {
	return &Command{name: "bzcat", codec: bzip2Codec, forceDecompress: true, forceStdout: true}
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.forceStdout {
		return "Decompress " + strings.TrimPrefix(c.codec.suffix, ".") + " data to standard output"
	}
	if c.forceDecompress {
		return "Decompress " + strings.TrimPrefix(c.codec.suffix, ".") + " files"
	}
	return "Compress or decompress files (" + strings.TrimPrefix(c.codec.suffix, ".") + ")"
}

type options struct {
	decompress bool
	stdout     bool
	keep       bool
	force      bool
}

// Run executes the applet.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(c.help())
	decompress := fs.BoolP("decompress", "d", false, "decompress")
	stdout := fs.BoolP("stdout", "c", false, "write to standard output and keep the input files")
	keep := fs.BoolP("keep", "k", false, "keep (don't delete) input files")
	force := fs.BoolP("force", "f", false, "force overwrite of the output file")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		decompress: c.forceDecompress || *decompress || c.codec.newWriter == nil,
		stdout:     c.forceStdout || *stdout,
		keep:       *keep,
		force:      *force,
	}

	files := fs.Args()
	if len(files) == 0 || (len(files) == 1 && files[0] == "-") {
		if err := c.transform(stdio.In, stdio.Out, opts.decompress); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
			return command.SilentFailure()
		}
		return nil
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

// processFile compresses or decompresses one file, either to stdout (-c) or in
// place with the suffix added/stripped.
func (c *Command) processFile(stdio command.IO, name string, opts options) error {
	if opts.stdout {
		in, err := os.Open(name) //nolint:gosec // user-named file
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		return c.transform(in, stdio.Out, opts.decompress)
	}

	out, err := c.outputName(name, opts.decompress)
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
	if err := c.transform(in, w, opts.decompress); err != nil {
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

// outputName derives the output filename: add the suffix when compressing, or
// strip it when decompressing.
func (c *Command) outputName(name string, decompress bool) (string, error) {
	if decompress {
		if !strings.HasSuffix(name, c.codec.suffix) {
			return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
		}
		return strings.TrimSuffix(name, c.codec.suffix), nil
	}
	return name + c.codec.suffix, nil
}

// transform copies r to w, compressing or decompressing with the codec.
func (c *Command) transform(r io.Reader, w io.Writer, decompress bool) error {
	if decompress {
		zr, err := c.codec.newReader(r)
		if err != nil {
			return err
		}
		defer func() { _ = zr.Close() }()
		_, err = io.Copy(w, zr) //nolint:gosec // decompressing user data
		return err
	}
	zw, err := c.codec.newWriter(w)
	if err != nil {
		return err
	}
	if _, err := io.Copy(zw, r); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func (c *Command) help() command.Help {
	h := command.Help{
		ExitStatus: "0  success.\n1  a file could not be read, written, or decoded.",
	}
	switch {
	case c.forceStdout:
		h.Description = "Decompress each FILE (" + c.codec.suffix + "), or standard input, to standard output."
		h.Examples = []command.Example{{Command: c.Name() + " file" + c.codec.suffix, Explain: "Write the decompressed data to standard output."}}
	case c.forceDecompress:
		h.Description = "Decompress each FILE in place, replacing FILE" + c.codec.suffix + " with FILE; with -c write to standard output instead."
		h.Examples = []command.Example{{Command: c.Name() + " file" + c.codec.suffix, Explain: "Decompress in place to 'file'."}}
	default:
		h.Description = "Compress each FILE in place to FILE" + c.codec.suffix + " (-d decompresses); with -c write to standard output. Reads standard input when no FILE is given."
		h.Examples = []command.Example{
			{Command: c.Name() + " file", Explain: "Compress 'file' to 'file" + c.codec.suffix + "'."},
			{Command: c.Name() + " -dc file" + c.codec.suffix, Explain: "Decompress to standard output."},
		}
	}
	return h
}
