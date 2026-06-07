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
type FlagSet struct {
	*pflag.FlagSet
	name    string
	usage   string
	help    *bool
	version *bool
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

// Parse parses args against the flag set. The returned proceed is false when
// the command should stop without doing work: either --help/--version was
// handled (err is nil) or parsing failed (err is a silent failure whose
// message has already been written to io.Err).
func (f *FlagSet) Parse(io IO, args []string) (proceed bool, err error) {
	if perr := f.FlagSet.Parse(args); perr != nil {
		fmt.Fprintf(io.Err, "%s: %v\n", f.name, perr)
		fmt.Fprintf(io.Err, "Try '%s --help' for more information.\n", f.name)
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

// WriteUsage writes a GNU-style usage block (a "Usage:" line followed by the
// option descriptions) to w.
func (f *FlagSet) WriteUsage(w io.Writer) {
	var b strings.Builder
	b.WriteString("Usage: ")
	b.WriteString(f.name)
	if f.usage != "" {
		b.WriteByte(' ')
		b.WriteString(f.usage)
	}
	b.WriteString("\n\nOptions:\n")
	b.WriteString(f.FlagUsages())
	_, _ = io.WriteString(w, b.String())
}
