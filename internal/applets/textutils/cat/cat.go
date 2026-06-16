// Package cat implements the cat applet: concatenate files (or standard input)
// to standard output, with the common GNU options.
package cat

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
)

// Command is the cat applet.
type Command struct{}

// New returns a cat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "cat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Concatenate files and print on the standard output" }

type options struct {
	number          bool
	numberNonBlank  bool
	squeeze         bool
	showEnds        bool
	showTabs        bool
	showNonprinting bool
}

// Run executes cat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Concatenate each FILE to standard output. With no FILE, or when FILE is\n" +
			"\"-\", read standard input. Options can number lines, squeeze repeated\n" +
			"blank lines, and make line ends or tabs visible.",
		Examples: []command.Example{
			{Command: "cat file.txt", Explain: "print file.txt to standard output"},
			{Command: "cat -n a.txt b.txt", Explain: "concatenate two files, numbering every line"},
			{Command: "cat", Explain: "copy standard input to standard output"},
		},
		ExitStatus: "0  success.\n1  one or more files could not be opened or read.",
	})
	number := fs.BoolP("number", "n", false, "number all output lines")
	numberNonBlank := fs.BoolP("number-nonblank", "b", false, "number non-empty output lines, overrides -n")
	squeeze := fs.BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	showEnds := fs.BoolP("show-ends", "E", false, "display $ at end of each line")
	showTabs := fs.BoolP("show-tabs", "T", false, "display TAB characters as ^I")
	showNonprinting := fs.BoolP("show-nonprinting", "v", false, "use ^ and M- notation, except for LFD and TAB")
	showAll := fs.BoolP("show-all", "A", false, "equivalent to -vET")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// -A is equivalent to -vET (show-nonprinting + show-ends + show-tabs).
	opts := options{
		number:          *number,
		numberNonBlank:  *numberNonBlank,
		squeeze:         *squeeze,
		showEnds:        *showEnds || *showAll,
		showTabs:        *showTabs || *showAll,
		showNonprinting: *showNonprinting || *showAll,
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	// Open every operand up front and stream the concatenation, so cat works in
	// constant memory on large files and pipes instead of reading everything in.
	// A failed open is reported and skipped; the others are still printed.
	var readers []io.Reader
	var closers []io.Closer
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "cat: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		readers = append(readers, r)
		closers = append(closers, r)
	}
	defer func() {
		for _, c := range closers {
			_ = c.Close()
		}
	}()

	src := io.Reader(io.MultiReader(readers...))

	// Apply -s/-T/-E/-v as a streaming filter before numbering, so the line
	// counter still sees the squeezed lines (matching the previous behavior).
	if opts.squeeze || opts.showTabs || opts.showEnds || opts.showNonprinting {
		raw := src // capture before reassigning src, or the goroutine would read the pipe itself
		pr, pw := io.Pipe()
		go func() { _ = pw.CloseWithError(renderStream(pw, raw, opts)) }()
		src = pr
	}

	if err := writeStream(stdio.Out, src, opts); err != nil {
		return command.Failure(err)
	}
	return firstErr
}

// renderStream copies r to w line by line, applying the display
// transformations (-s squeeze blank lines, -T show tabs, -E show line ends).
func renderStream(w io.Writer, r io.Reader, opts options) error {
	br := bufio.NewReader(r)
	prevBlank := false
	for {
		chunk, err := br.ReadString('\n')
		if chunk != "" {
			body, nl := chunk, ""
			if strings.HasSuffix(chunk, "\n") {
				body, nl = chunk[:len(chunk)-1], "\n"
			}
			emit := true
			if opts.squeeze {
				blank := body == ""
				if blank && prevBlank {
					emit = false
				}
				prevBlank = blank
			}
			if emit {
				// -v renders non-printing bytes with caret/M- notation. It
				// deliberately leaves TAB untouched (only -T converts TAB to
				// ^I) and never touches the trailing newline.
				if opts.showNonprinting {
					body = showNonprinting(body)
				}
				if opts.showTabs {
					body = strings.ReplaceAll(body, "\t", "^I")
				}
				if opts.showEnds {
					body += "$"
				}
				if _, werr := io.WriteString(w, body+nl); werr != nil {
					return werr
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// showNonprinting renders non-printing bytes the way GNU cat -v does:
//
//   - Control characters 0x00-0x1F (except TAB 0x09 and LF 0x0A, which GNU cat
//     leaves alone for -v) become caret notation: ^@ ... ^_ (^X = 0x40+byte).
//   - DEL (0x7F) becomes ^?.
//   - Bytes 0x80-0xFF get the M- prefix, followed by the caret/printable form
//     of the low 7 bits: 0x80-0x9F -> M-^@ ... M-^_, 0xA0-0xFE -> M-<char>,
//     0xFF -> M-^?.
//
// TAB and LF are intentionally passed through untouched; only -T converts TAB.
// Input is processed byte by byte (not rune by rune), matching GNU semantics.
func showNonprinting(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\t' || c == '\n':
			b.WriteByte(c)
		case c < 0x20: // other C0 control chars
			b.WriteByte('^')
			b.WriteByte(c + 0x40)
		case c < 0x7f: // printable ASCII
			b.WriteByte(c)
		case c == 0x7f: // DEL
			b.WriteString("^?")
		default: // c >= 0x80: meta (high-bit) byte
			b.WriteString("M-")
			m := c & 0x7f
			switch {
			case m < 0x20:
				b.WriteByte('^')
				b.WriteByte(m + 0x40)
			case m == 0x7f:
				b.WriteString("^?")
			default:
				b.WriteByte(m)
			}
		}
	}
	return b.String()
}

// writeStream writes src to w, numbering lines when -n/-b is set (streaming via
// the Numberer) or copying through otherwise.
func writeStream(w io.Writer, src io.Reader, opts options) error {
	if opts.number || opts.numberNonBlank {
		n := textproc.Numberer{
			Style:     numberStyle(opts.numberNonBlank),
			Start:     1,
			Increment: 1,
			Width:     6,
			Separator: "\t",
		}
		return n.WriteTo(w, src)
	}
	_, err := io.Copy(w, src)
	return err
}

func numberStyle(nonBlank bool) textproc.NumberStyle {
	if nonBlank {
		return textproc.NumberNonBlank
	}
	return textproc.NumberAll
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
