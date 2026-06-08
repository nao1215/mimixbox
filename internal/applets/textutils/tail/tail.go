// Package tail implements the tail applet: print the last part of files (or
// standard input).
package tail

import (
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/textproc"
)

// Command is the tail applet.
type Command struct{}

// New returns a tail command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tail" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the last NUMBER(default=10) lines" }

// Run executes tail.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [FILE]...", stdio.Err)
	lines := fs.IntP("lines", "n", 10, "output the last NUM lines instead of the last 10")
	bytesN := fs.IntP("bytes", "c", 0, "output the last NUM bytes of each file")
	quiet := fs.BoolP("quiet", "q", false, "never print headers giving file names")
	verbose := fs.BoolP("verbose", "v", false, "always print headers giving file names")
	followMode := fs.StringP("follow", "f", "", "output appended data as the file grows; MODE is 'name' or 'descriptor'")
	fs.Lookup("follow").NoOptDefVal = "descriptor"
	followName := fs.BoolP("follow-name", "F", false, "same as --follow=name --retry")
	retry := fs.Bool("retry", false, "keep trying to open a file even if it is inaccessible")
	sleepInterval := fs.Float64P("sleep-interval", "s", 1.0, "seconds to wait between iterations when following")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *sleepInterval <= 0 {
		return command.Failuref("invalid number of seconds: %g", *sleepInterval)
	}
	if fs.Changed("follow") && *followMode != "name" && *followMode != "descriptor" {
		return command.Failuref("invalid argument %q for '--follow'; valid arguments are 'name', 'descriptor'", *followMode)
	}

	following := fs.Changed("follow") || *followName
	reopen := *followName || *followMode == "name"
	retryOpen := *retry || *followName

	files := fs.Args()
	if len(files) == 0 {
		files = []string{"-"}
	}
	showHeader := (len(files) > 1 || *verbose) && !*quiet

	var firstErr error
	for i, name := range files {
		r, err := command.Open(stdio, name)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tail: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
			continue
		}
		if showHeader {
			writeHeader(stdio.Out, name, i == 0)
		}
		if *bytesN > 0 {
			err = textproc.TailBytes(stdio.Out, r, *bytesN)
		} else {
			err = textproc.TailLines(stdio.Out, r, *lines)
		}
		_ = r.Close()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "tail: %s\n", command.FileError(name, err))
			firstErr = keep(firstErr)
		}
	}

	if following {
		// Standard input cannot be polled for growth, so only real files are
		// followed. With -F/--retry a not-yet-existing file is still tracked.
		paths := followablePaths(files)
		targets := newFollowTargets(paths, retryOpen)
		defer closeAll(targets)
		follow(ctx, stdio, targets, *sleepInterval, reopen, showHeader)
	}
	return firstErr
}

// followablePaths returns the file operands that can be polled for growth,
// dropping the "-" (standard input) pseudo-file.
func followablePaths(files []string) []string {
	paths := make([]string, 0, len(files))
	for _, name := range files {
		if name == "-" {
			continue
		}
		paths = append(paths, name)
	}
	return paths
}

func writeHeader(w io.Writer, name string, first bool) {
	label := name
	if name == "-" {
		label = "standard input"
	}
	if first {
		_, _ = fmt.Fprintf(w, "==> %s <==\n", label)
		return
	}
	_, _ = fmt.Fprintf(w, "\n==> %s <==\n", label)
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
