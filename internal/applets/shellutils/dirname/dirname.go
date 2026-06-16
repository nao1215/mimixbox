// Package dirname implements the dirname applet: strip the last component from
// each file name.
package dirname

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the dirname applet.
type Command struct{}

// New returns a dirname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dirname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print only directory path" }

// Run executes dirname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... NAME...", stdio.Err).WithHelp(command.Help{
		Description: "Print each NAME with its last non-slash component and any trailing slashes removed; " +
			"if NAME contains no slashes, print \".\" (meaning the current directory).",
		Examples: []command.Example{
			{Command: "dirname /usr/bin/sort", Explain: "Print \"/usr/bin\"."},
			{Command: "dirname stdio.h", Explain: "Print \".\"."},
			{Command: "dirname /usr/lib /tmp/x", Explain: "Print the directory of each operand on its own line."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. no operand was given).",
	})
	zero := fs.BoolP("zero", "z", false, "end each output line with NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "dirname: missing operand")
		return command.SilentFailure()
	}

	end := byte('\n')
	if *zero {
		end = 0
	}
	for _, name := range names {
		_, _ = fmt.Fprintf(stdio.Out, "%s%c", dir(name), end)
	}
	return nil
}

// dir returns the directory part of p, matching GNU dirname: trailing slashes
// are ignored, a path with no slash yields ".", and a path of only slashes
// yields "/".
func dir(p string) string {
	trimmed := strings.TrimRight(p, "/")
	if trimmed == "" {
		if p == "" {
			// No slash at all.
			return "."
		}
		// The path was made entirely of slashes.
		return "/"
	}
	i := strings.LastIndexByte(trimmed, '/')
	switch {
	case i < 0:
		return "."
	case i == 0:
		return "/"
	default:
		return strings.TrimRight(trimmed[:i], "/")
	}
}
