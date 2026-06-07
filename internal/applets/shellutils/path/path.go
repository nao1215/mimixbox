// Package path implements the path applet: a MimixBox-original command that
// manipulates a filename path, extracting its directory, basename, extension,
// canonical form, or absolute form.
package path

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the path applet.
type Command struct{}

// New returns a path command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "path" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Manipulate filename path" }

// Run executes path.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... PATH...", stdio.Err)
	abs := fs.BoolP("absolute", "a", false, "Print absolute path")
	base := fs.BoolP("basename", "b", false, "Print basename (filename)")
	canonical := fs.BoolP("canonical", "c", false, "Print canonical path (default)")
	dir := fs.BoolP("dirname", "d", false, "Print path without filename")
	ext := fs.BoolP("extension", "e", false, "Print file extention")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		fmt.Fprintln(stdio.Err, "path: missing operand")
		return command.SilentFailure()
	}
	p := names[0]

	allOff := !*abs && !*base && !*canonical && !*dir && !*ext

	if *abs {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return command.Failuref("path: can't get absolute path")
		}
		fmt.Fprintf(stdio.Out, "%s\n", absPath)
	}
	if *base {
		fmt.Fprintf(stdio.Out, "%s\n", filepath.Base(p))
	}
	if *canonical || allOff {
		fmt.Fprintf(stdio.Out, "%s\n", filepath.Clean(p))
	}
	if *dir {
		fmt.Fprintf(stdio.Out, "%s\n", filepath.Dir(p))
	}
	if *ext {
		fmt.Fprintf(stdio.Out, "%s\n", filepath.Ext(p))
	}

	return nil
}
