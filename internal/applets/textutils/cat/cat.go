// Package cat implements the cat applet: concatenate files (or standard input)
// to standard output, with the common GNU options.
package cat

import (
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
	number         bool
	numberNonBlank bool
	squeeze        bool
	showEnds       bool
	showTabs       bool
}

// Run executes cat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	number := fs.BoolP("number", "n", false, "number all output lines")
	numberNonBlank := fs.BoolP("number-nonblank", "b", false, "number non-empty output lines, overrides -n")
	squeeze := fs.BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	showEnds := fs.BoolP("show-ends", "E", false, "display $ at end of each line")
	showTabs := fs.BoolP("show-tabs", "T", false, "display TAB characters as ^I")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	opts := options{
		number:         *number,
		numberNonBlank: *numberNonBlank,
		squeeze:        *squeeze,
		showEnds:       *showEnds,
		showTabs:       *showTabs,
	}

	text, readErr := concat(stdio, fs.Args())
	if err := write(stdio.Out, render(text, opts), opts); err != nil {
		return command.Failure(err)
	}
	return readErr
}

// concat reads every operand (defaulting to standard input when there are none)
// and returns the joined text. A failed open or read is reported on stderr but
// does not stop the remaining files; the returned error only sets the exit
// code, because its message was already printed.
func concat(stdio command.IO, files []string) (string, error) {
	if len(files) == 0 {
		files = []string{"-"}
	}
	var b strings.Builder
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			fmt.Fprintf(stdio.Err, "cat: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		_, err = io.Copy(&b, r)
		_ = r.Close()
		if err != nil {
			fmt.Fprintf(stdio.Err, "cat: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}
	return b.String(), firstErr
}

// render applies the display transformations (-s, -T, -E) and returns new text.
// Numbering is applied later by write so its counter sees the squeezed lines.
func render(text string, opts options) string {
	if !opts.squeeze && !opts.showTabs && !opts.showEnds {
		return text
	}
	lines := splitKeepNewline(text)
	if opts.squeeze {
		lines = squeezeBlank(lines)
	}
	var b strings.Builder
	for _, l := range lines {
		body := l.body
		if opts.showTabs {
			body = strings.ReplaceAll(body, "\t", "^I")
		}
		if opts.showEnds {
			body += "$"
		}
		b.WriteString(body)
		b.WriteString(l.newline)
	}
	return b.String()
}

func write(w io.Writer, text string, opts options) error {
	if opts.number || opts.numberNonBlank {
		n := textproc.Numberer{
			Style:     numberStyle(opts.numberNonBlank),
			Start:     1,
			Increment: 1,
			Width:     6,
			Separator: "\t",
		}
		return n.WriteTo(w, strings.NewReader(text))
	}
	_, err := io.WriteString(w, text)
	return err
}

func numberStyle(nonBlank bool) textproc.NumberStyle {
	if nonBlank {
		return textproc.NumberNonBlank
	}
	return textproc.NumberAll
}

type line struct {
	body    string
	newline string
}

func splitKeepNewline(s string) []line {
	var lines []line
	for s != "" {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			lines = append(lines, line{body: s})
			break
		}
		lines = append(lines, line{body: s[:i], newline: "\n"})
		s = s[i+1:]
	}
	return lines
}

func squeezeBlank(lines []line) []line {
	out := make([]line, 0, len(lines))
	prevBlank := false
	for _, l := range lines {
		blank := l.body == ""
		if blank && prevBlank {
			continue
		}
		prevBlank = blank
		out = append(out, l)
	}
	return out
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
