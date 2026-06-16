// Package nl implements the nl applet: write files (or standard input) to
// standard output with line numbers added.
package nl

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
)

// Command is the nl applet.
type Command struct{}

// New returns an nl command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nl" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Write each FILE to standard output with line numbers added"
}

// Run executes nl.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err).WithHelp(command.Help{
		Description: "Write each FILE (or standard input when no FILE is given, or FILE is '-') to standard output with line numbers prepended. Input is divided into header, body and footer sections by the delimiter lines \\:\\:\\:, \\:\\: and \\:; each section is numbered with its own STYLE (-h, -b, -f). By default only non-blank body lines are numbered. A STYLE is one of a (all lines), t (non-empty lines), n (no lines) or pBRE (lines matching the basic regular expression BRE).",
		Examples: []command.Example{
			{Command: "nl file.txt", Explain: "Number the non-blank lines of file.txt."},
			{Command: "nl -b a -w 4 file.txt", Explain: "Number every line in a 4-column field."},
			{Command: "nl -h a -f a file.txt", Explain: "Also number every header and footer line."},
		},
		ExitStatus: "0  success.\n1  a file could not be read.",
	})
	body := fs.StringP("body-numbering", "b", "t", "use STYLE for numbering body lines (a, t, n, or pBRE)")
	header := fs.StringP("header-numbering", "h", "n", "use STYLE for numbering header lines (a, t, n, or pBRE)")
	footer := fs.StringP("footer-numbering", "f", "n", "use STYLE for numbering footer lines (a, t, n, or pBRE)")
	join := fs.IntP("join-blank-lines", "l", 1, "group of NUMBER empty lines counted as one")
	separator := fs.StringP("number-separator", "s", "\t", "add STRING after possible line number")
	width := fs.IntP("number-width", "w", 6, "use NUMBER columns for line numbers")
	format := fs.StringP("number-format", "n", "rn", "insert line numbers according to FORMAT (ln, rn, rz)")
	start := fs.IntP("starting-line-number", "v", 1, "first line number for each section")
	increment := fs.IntP("line-increment", "i", 1, "line number increment at each line")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	bodyStyle, bodyRe, ok := sectionStyle(*body)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "nl: invalid body numbering style: %q\n", *body)
		return command.SilentFailure()
	}
	headerStyle, headerRe, ok := sectionStyle(*header)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "nl: invalid header numbering style: %q\n", *header)
		return command.SilentFailure()
	}
	footerStyle, footerRe, ok := sectionStyle(*footer)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "nl: invalid footer numbering style: %q\n", *footer)
		return command.SilentFailure()
	}
	if *join < 1 {
		_, _ = fmt.Fprintf(stdio.Err, "nl: invalid line number of blank lines: %d\n", *join)
		return command.SilentFailure()
	}
	justify, ok := numberFormat(*format)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "nl: invalid line numbering format: %q\n", *format)
		return command.SilentFailure()
	}

	numberer := textproc.Numberer{
		Style:          bodyStyle,
		Pattern:        bodyRe,
		Start:          *start,
		Increment:      *increment,
		Width:          *width,
		Separator:      *separator,
		Justify:        justify,
		PadBlank:       true,
		Sections:       true,
		HeaderStyle:    headerStyle,
		HeaderRegexp:   headerRe,
		FooterStyle:    footerStyle,
		FooterRegexp:   footerRe,
		JoinBlankLines: *join,
	}

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}

	// Open every operand and number the concatenation as a single stream, so
	// the line counter spans file boundaries (as nl does) without ever holding
	// the whole input in memory.
	var readers []io.Reader
	var closers []io.Closer
	var firstErr error
	for _, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "nl: %s\n", command.FileError(name, err))
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

	if err := numberer.WriteTo(stdio.Out, io.MultiReader(readers...)); err != nil {
		return command.Failure(err)
	}
	return firstErr
}

// sectionStyle parses an nl numbering STYLE: "a" (all lines), "t" (non-empty
// lines), "n" (no lines) or "pBRE" (lines matching the basic regular expression
// BRE). For pBRE it returns the compiled pattern.
func sectionStyle(s string) (textproc.NumberStyle, *regexp.Regexp, bool) {
	switch s {
	case "a":
		return textproc.NumberAll, nil, true
	case "t":
		return textproc.NumberNonBlank, nil, true
	case "n":
		return textproc.NumberNone, nil, true
	default:
		if strings.HasPrefix(s, "p") && len(s) > 1 {
			re, err := regexp.Compile(s[1:])
			if err != nil {
				return 0, nil, false
			}
			return textproc.NumberRegexp, re, true
		}
		return 0, nil, false
	}
}

func numberFormat(s string) (textproc.NumberJustify, bool) {
	switch s {
	case "ln":
		return textproc.JustifyLeft, true
	case "rn":
		return textproc.JustifyRight, true
	case "rz":
		return textproc.JustifyRightZero, true
	default:
		return 0, false
	}
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
