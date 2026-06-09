package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/nao1215/mimixbox/internal/version"
	"github.com/spf13/pflag"
)

// FlagSet wraps pflag.FlagSet with the conventions every MimixBox command
// shares: GNU-style parsing (long --options, clustered -short flags, "--" to
// end options, and operands interspersed with options), plus the standard
// --help and --version flags. Using pflag is what lets a MimixBox applet stand
// in for the system command of the same name.
// Example is one worked example shown in --help: the command line and a short
// explanation of what it does.
type Example struct {
	Command string
	Explain string
}

// Help holds the optional, self-describing sections an applet can attach to its
// --help output so that a human or an LLM can understand the command without
// prior knowledge: a one-paragraph description, worked examples, an exit-status
// summary, and compatibility notes. All fields are optional; empty sections are
// not rendered.
type Help struct {
	Description string
	Examples    []Example
	ExitStatus  string
	Notes       []string
}

type FlagSet struct {
	*pflag.FlagSet
	name    string
	usage   string
	help    *bool
	version *bool
	doc     Help
}

// NewFlagSet returns a FlagSet for command name. usage is the operand summary
// shown after the command name, e.g. "[OPTION]... [FILE]...".
func NewFlagSet(name, usage string, stderr io.Writer) *FlagSet {
	pf := pflag.NewFlagSet(name, pflag.ContinueOnError)
	pf.SetInterspersed(true)
	pf.SetOutput(stderr)
	// Silence pflag's own usage; FlagSet renders GNU-style messages itself.
	pf.Usage = func() {}

	f := &FlagSet{FlagSet: pf, name: name, usage: usage}
	f.help = pf.Bool("help", false, "display this help and exit")
	f.version = pf.Bool("version", false, "output version information and exit")
	return f
}

// WithHelp attaches self-describing help sections (description, examples, exit
// status, notes) that WriteUsage renders below the options. It returns the
// FlagSet so it can be chained onto NewFlagSet.
func (f *FlagSet) WithHelp(h Help) *FlagSet {
	f.doc = h
	return f
}

// Parse parses args against the flag set. The returned proceed is false when
// the command should stop without doing work: either --help/--version was
// handled (err is nil) or parsing failed (err is a silent failure whose
// message has already been written to io.Err).
func (f *FlagSet) Parse(io IO, args []string) (proceed bool, err error) {
	if perr := f.FlagSet.Parse(args); perr != nil {
		_, _ = fmt.Fprintf(io.Err, "%s: %v\n", f.name, perr)
		_, _ = fmt.Fprintf(io.Err, "Try '%s --help' for more information.\n", f.name)
		return false, SilentFailure()
	}
	if *f.help {
		f.WriteUsage(io.Out)
		return false, nil
	}
	if *f.version {
		version.Print(io.Out, f.name)
		return false, nil
	}
	return true, nil
}

// WriteUsage writes a GNU-style usage block to w: a "Usage:" line, an optional
// one-paragraph description, the option descriptions, and any optional
// Examples / Exit status / Notes sections attached via WithHelp.
func (f *FlagSet) WriteUsage(w io.Writer) {
	var b strings.Builder
	b.WriteString("Usage: ")
	b.WriteString(f.name)
	if f.usage != "" {
		b.WriteByte(' ')
		b.WriteString(f.usage)
	}
	b.WriteByte('\n')

	if f.doc.Description != "" {
		b.WriteByte('\n')
		b.WriteString(f.doc.Description)
		b.WriteByte('\n')
	}

	b.WriteString("\nOptions:\n")
	b.WriteString(f.FlagUsages())

	if len(f.doc.Examples) > 0 {
		b.WriteString("\nExamples:\n")
		width := 0
		for _, ex := range f.doc.Examples {
			if len(ex.Command) > width {
				width = len(ex.Command)
			}
		}
		for _, ex := range f.doc.Examples {
			fmt.Fprintf(&b, "  %-*s", width, ex.Command)
			if ex.Explain != "" {
				b.WriteString("  ")
				b.WriteString(ex.Explain)
			}
			b.WriteByte('\n')
		}
	}

	if f.doc.ExitStatus != "" {
		b.WriteString("\nExit status:\n")
		b.WriteString(indentLines(f.doc.ExitStatus))
	}

	if len(f.doc.Notes) > 0 {
		b.WriteString("\nNotes:\n")
		for _, note := range f.doc.Notes {
			b.WriteString("  - ")
			b.WriteString(note)
			b.WriteByte('\n')
		}
	}

	_, _ = io.WriteString(w, b.String())
}

// indentLines indents every line of s with two spaces, ensuring a trailing
// newline so sections stack cleanly.
func indentLines(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
