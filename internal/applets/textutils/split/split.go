// Package split implements the split applet: break an input file into smaller
// output files by a line count or a byte count.
package split

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the split applet.
type Command struct{}

// New returns a split command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "split" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Split a file into pieces" }

// Run executes split.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [INPUT [PREFIX]]", stdio.Err).WithHelp(command.Help{
		Description: "Split INPUT into fixed-size output files named PREFIXaa, PREFIXab, and so on " +
			"(PREFIX defaults to x). With no INPUT, or when INPUT is -, read standard input.",
		Examples: []command.Example{
			{Command: "split -l 100 big.txt part_", Explain: "Split big.txt into files of 100 lines each, named part_aa, part_ab, ..."},
			{Command: "split -b 1M data.bin", Explain: "Split data.bin into 1 MiB pieces named xaa, xab, ..."},
			{Command: "split -d -a 3 big.txt part_", Explain: "Use 3-digit numeric suffixes: part_000, part_001, ..."},
			{Command: "split -l 100 --additional-suffix=.txt big.txt part_", Explain: "Append .txt to each output name: part_aa.txt, part_ab.txt, ..."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. the input could not be read or an option was invalid).",
	})
	lines := fs.IntP("lines", "l", 1000, "put NUMBER lines per output file")
	byteSpec := fs.StringP("bytes", "b", "", "put SIZE bytes per output file (suffixes K, M allowed)")
	// numeric-suffixes is a string so it can accept an optional =FROM value while
	// still serving as the -d shorthand. NoOptDefVal makes the bare flag (-d or
	// --numeric-suffixes) mean "start at 0".
	numericSpec := fs.StringP("numeric-suffixes", "d", "", "use numeric suffixes starting at FROM (default 0)")
	fs.Lookup("numeric-suffixes").NoOptDefVal = "0"
	addSuffix := fs.String("additional-suffix", "", "append an additional SUFFIX to file names")
	suffixLen := fs.IntP("suffix-length", "a", 2, "generate suffixes of length N")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	input := "-"
	prefix := "x"
	if len(rest) > 0 {
		input = rest[0]
	}
	if len(rest) > 1 {
		prefix = rest[1]
	}

	naming, nerr := newNamingScheme(fs.Changed("numeric-suffixes"), *numericSpec, *addSuffix, *suffixLen)
	if nerr != nil {
		_, _ = fmt.Fprintf(stdio.Err, "split: %v\n", nerr)
		return command.SilentFailure()
	}

	r, err := command.Open(stdio, input)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "split: %s\n", command.FileError(input, err))
		return command.SilentFailure()
	}
	defer func() { _ = r.Close() }()

	if *byteSpec != "" {
		size, perr := parseSize(*byteSpec)
		if perr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "split: invalid number of bytes: %q\n", *byteSpec)
			return command.SilentFailure()
		}
		return c.byBytes(stdio, r, prefix, size, naming)
	}
	if *lines <= 0 {
		_, _ = fmt.Fprintf(stdio.Err, "split: invalid number of lines: %d\n", *lines)
		return command.SilentFailure()
	}
	return c.byLines(stdio, r, prefix, *lines, naming)
}

// namingScheme describes how successive output file names are built from the
// index: alphabetic (aa, ab, ...) or numeric (00, 01, ...), the suffix width,
// and an optional trailing suffix appended after the index part.
type namingScheme struct {
	numeric   bool
	from      int
	addSuffix string
	suffixLen int
}

// newNamingScheme validates and assembles a namingScheme from the parsed flags.
// changed reports whether --numeric-suffixes/-d was supplied; spec is its
// (optional) =FROM value.
func newNamingScheme(changed bool, spec, addSuffix string, suffixLen int) (namingScheme, error) {
	if suffixLen <= 0 {
		return namingScheme{}, fmt.Errorf("invalid suffix length: %d", suffixLen)
	}
	ns := namingScheme{addSuffix: addSuffix, suffixLen: suffixLen}
	if changed {
		ns.numeric = true
		if spec != "" {
			from, err := strconv.Atoi(spec)
			if err != nil || from < 0 {
				return namingScheme{}, fmt.Errorf("invalid start value for numeric suffixes: %q", spec)
			}
			ns.from = from
		}
	}
	return ns, nil
}

// name returns the output file name for the idx-th piece (idx counts from 0).
func (ns namingScheme) name(prefix string, idx int) string {
	return prefix + ns.suffixPart(idx) + ns.addSuffix
}

// suffixPart returns the index portion of the file name (without prefix or the
// additional suffix), padded to suffixLen.
func (ns namingScheme) suffixPart(idx int) string {
	if ns.numeric {
		return fmt.Sprintf("%0*d", ns.suffixLen, ns.from+idx)
	}
	return alphaSuffix(idx, ns.suffixLen)
}

// parseSize parses a byte count that may carry a K (1024) or M (1048576) suffix.
func parseSize(s string) (int, error) {
	mult := 1
	switch {
	case strings.HasSuffix(s, "K"):
		mult, s = 1024, strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		mult, s = 1024*1024, strings.TrimSuffix(s, "M")
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid size")
	}
	return n * mult, nil
}

// byLines writes perFile lines into each successive output file.
func (c *Command) byLines(stdio command.IO, r io.Reader, prefix string, perFile int, ns namingScheme) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	idx, count := 0, 0
	var w *os.File
	closeCur := func() error {
		if w != nil {
			err := w.Close()
			w = nil
			return err
		}
		return nil
	}
	for sc.Scan() {
		if w == nil {
			f, err := create(ns, prefix, idx)
			if err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "split: %v\n", err)
				return command.SilentFailure()
			}
			w = f
			idx++
		}
		if _, err := w.WriteString(sc.Text() + "\n"); err != nil {
			_ = closeCur()
			return command.Failure(err)
		}
		count++
		if count == perFile {
			if err := closeCur(); err != nil {
				return command.Failure(err)
			}
			count = 0
		}
	}
	if err := closeCur(); err != nil {
		return command.Failure(err)
	}
	return sc.Err()
}

// byBytes writes perFile bytes into each successive output file.
func (c *Command) byBytes(stdio command.IO, r io.Reader, prefix string, perFile int, ns namingScheme) error {
	br := bufio.NewReader(r)
	buf := make([]byte, perFile)
	idx := 0
	for {
		n, err := io.ReadFull(br, buf)
		if n > 0 {
			f, cerr := create(ns, prefix, idx)
			if cerr != nil {
				_, _ = fmt.Fprintf(stdio.Err, "split: %v\n", cerr)
				return command.SilentFailure()
			}
			idx++
			if _, werr := f.Write(buf[:n]); werr != nil {
				_ = f.Close()
				return command.Failure(werr)
			}
			if cerr := f.Close(); cerr != nil {
				return command.Failure(cerr)
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return command.Failure(err)
		}
	}
}

// create opens the idx-th output file, named according to the naming scheme.
func create(ns namingScheme, prefix string, idx int) (*os.File, error) {
	name := ns.name(prefix, idx)
	return os.Create(name) //nolint:gosec // writing a user-named output file is the point
}

// alphaSuffix returns the alphabetic split suffix for the idx-th file padded to
// width characters (aa, ab, ..., az, ba, ...). It mirrors GNU's base-26 scheme:
// the least-significant position varies fastest.
func alphaSuffix(idx, width int) string {
	b := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		b[i] = byte('a' + idx%26)
		idx /= 26
	}
	return string(b)
}
