// Package basename implements the basename applet: strip directory and an
// optional suffix from file names.
package basename

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the basename applet.
type Command struct{}

// New returns a basename command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "basename" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print basename (PATH without \"/\") from file path" }

// Run executes basename.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NAME [SUFFIX] | OPTION... NAME...", stdio.Err)
	multiple := fs.BoolP("multiple", "a", false, "support multiple arguments and treat each as a NAME")
	suffix := fs.StringP("suffix", "s", "", "remove a trailing SUFFIX; implies -a")
	zero := fs.BoolP("zero", "z", false, "end each output line with NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		fmt.Fprintln(stdio.Err, "basename: missing operand")
		return command.SilentFailure()
	}

	suf := *suffix
	multi := *multiple || *suffix != ""
	if !multi {
		// Classic form: basename NAME [SUFFIX].
		if len(names) > 2 {
			fmt.Fprintf(stdio.Err, "basename: extra operand '%s'\n", names[2])
			return command.SilentFailure()
		}
		if len(names) == 2 {
			suf = names[1]
		}
		names = names[:1]
	}

	end := byte('\n')
	if *zero {
		end = 0
	}
	for _, name := range names {
		fmt.Fprintf(stdio.Out, "%s%c", stripSuffix(base(name), suf), end)
	}
	return nil
}

// base returns the final path component, matching GNU basename: trailing
// slashes are removed first, and a path made only of slashes becomes "/".
func base(p string) string {
	if p == "" {
		return ""
	}
	trimmed := strings.TrimRight(p, "/")
	if trimmed == "" {
		// The original was made entirely of slashes.
		return "/"
	}
	if i := strings.LastIndexByte(trimmed, '/'); i >= 0 {
		return trimmed[i+1:]
	}
	return trimmed
}

// stripSuffix removes suffix from name unless that would leave nothing or the
// name equals the suffix, matching GNU behaviour.
func stripSuffix(name, suffix string) string {
	if suffix == "" || name == suffix {
		return name
	}
	return strings.TrimSuffix(name, suffix)
}
