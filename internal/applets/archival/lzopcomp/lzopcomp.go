// Package lzopcomp implements the lzop family of applets (lzop, unlzop and
// lzopcat). lzop compresses and decompresses files in the ".lzo" container
// format, which frames LZO1X-compressed blocks with per-file and per-block
// metadata. The LZO1X codec itself comes from github.com/rasky/go-lzo; this
// package adds the lzop container (magic, file header with header checksum, and
// length-prefixed compressed blocks with Adler-32 checksums) so the output is
// interoperable with the upstream lzop utility.
//
// The three applets share one file-handling model (stdin/stdout, -c, -k, -f,
// in-place ".lzo" suffix rename), so each is just a small configuration of that
// model, mirroring the xzcomp package.
package lzopcomp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/archival/comp"
	"github.com/nao1215/mimixbox/internal/command"
)

// suffix is the filename extension lzop appends to compressed files.
const suffix = ".lzo"

// Command is one lzop-family applet.
type Command struct {
	name            string
	forceDecompress bool // unlzop/lzopcat always decompress
	forceStdout     bool // lzopcat always writes to stdout
}

// NewLzop returns the lzop applet (compresses by default).
func NewLzop() *Command { return &Command{name: "lzop"} }

// NewUnlzop returns the unlzop applet (lzop -d).
func NewUnlzop() *Command { return &Command{name: "unlzop", forceDecompress: true} }

// NewLzopcat returns the lzopcat applet (lzop -dc).
func NewLzopcat() *Command {
	return &Command{name: "lzopcat", forceDecompress: true, forceStdout: true}
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch {
	case c.forceStdout:
		return "Decompress lzop (.lzo) data to standard output"
	case c.forceDecompress:
		return "Decompress lzop (.lzo) files"
	default:
		return "Compress or decompress files (.lzo)"
	}
}

// Run executes the applet by delegating the shared file-handling model to the
// comp frontend; this Command only supplies the codec, the -t test mode and the
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
		Decompress: c.forceDecompress || *decompress || *test,
		Stdout:     c.forceStdout || *stdout,
		Keep:       *keep,
		Force:      *force,
		Test:       *test,
	}
	return cfg.Run(stdio, opts, fs.Args())
}

// outputName derives the output filename: add .lzo when compressing, strip it
// when decompressing.
func outputName(name string, decompress bool) (string, error) {
	if decompress {
		switch {
		case strings.HasSuffix(name, suffix):
			return strings.TrimSuffix(name, suffix), nil
		case strings.HasSuffix(name, ".tzo"):
			return strings.TrimSuffix(name, ".tzo") + ".tar", nil
		default:
			return "", fmt.Errorf("%s: unknown suffix -- ignored", name)
		}
	}
	return name + suffix, nil
}

// transform copies r to w, lzop-compressing or -decompressing along the way.
func transform(r io.Reader, w io.Writer, decompress bool) error {
	if decompress {
		return decompressStream(r, w)
	}
	bw := bufio.NewWriter(w)
	if err := compressStream(r, bw); err != nil {
		return err
	}
	return bw.Flush()
}

// testStream verifies that r is a valid lzop stream by decompressing it and
// discarding the output.
func testStream(r io.Reader) error {
	return decompressStream(r, io.Discard)
}

func (c *Command) help() command.Help {
	h := command.Help{
		ExitStatus: "0  success.\n1  a file could not be read, written, or decoded.",
		Notes: []string{
			"Uses the LZO1X codec inside the standard lzop (.lzo) container, so output is compatible with the upstream lzop utility.",
		},
	}
	switch {
	case c.forceStdout:
		h.Description = "Decompress each FILE (.lzo), or standard input, to standard output."
		h.Examples = []command.Example{{Command: "lzopcat file.lzo", Explain: "Write the decompressed data to standard output."}}
	case c.forceDecompress:
		h.Description = "Decompress each FILE in place, replacing FILE.lzo with FILE; with -c write to standard output instead."
		h.Examples = []command.Example{{Command: "unlzop file.lzo", Explain: "Decompress in place to 'file'."}}
	default:
		h.Description = "Compress each FILE in place to FILE.lzo (-d decompresses); with -c write to standard output. Reads standard input when no FILE is given."
		h.Examples = []command.Example{
			{Command: "lzop file", Explain: "Compress 'file' to 'file.lzo'."},
			{Command: "lzop -dc file.lzo", Explain: "Decompress to standard output."},
			{Command: "lzop -t file.lzo", Explain: "Test the integrity of 'file.lzo'."},
		}
	}
	return h
}
