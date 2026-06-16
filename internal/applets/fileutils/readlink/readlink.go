// Package readlink implements the readlink applet: print the target of a
// symbolic link, or with -f the canonicalized absolute path.
package readlink

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the readlink applet.
type Command struct{}

// New returns a readlink command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "readlink" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print resolved symbolic links or canonical file names" }

// Run executes readlink.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the target of each symbolic link FILE. With -f, print the canonicalized " +
			"absolute path, following every symlink in the chain.",
		Examples: []command.Example{
			{Command: "readlink /usr/bin/vi", Explain: "Print the immediate target of the symlink."},
			{Command: "readlink -f /usr/bin/vi", Explain: "Print the fully resolved canonical path."},
			{Command: "readlink -n link", Explain: "Print the target without a trailing newline."},
		},
		ExitStatus: "0  every operand was resolved.\n1  an operand was not a symlink or could not be resolved.",
	})
	canon := fs.BoolP("canonicalize", "f", false, "canonicalize by following every symlink")
	quiet := fs.BoolP("no-newline", "n", false, "do not output the trailing newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing operand")
	}

	var firstErr error
	for _, name := range files {
		target, err := resolve(name, *canon)
		if err != nil {
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		end := "\n"
		if *quiet {
			end = ""
		}
		if _, werr := io.WriteString(stdio.Out, target+end); werr != nil {
			return command.Failure(werr)
		}
	}
	return firstErr
}

// resolve returns the link target for name. With canon set it returns the
// fully canonicalized absolute path; otherwise it reads the link directly,
// which fails when name is not a symlink (matching GNU readlink).
func resolve(name string, canon bool) (string, error) {
	if canon {
		abs, err := filepath.Abs(name)
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(abs)
	}
	target, err := os.Readlink(name)
	if err != nil {
		return "", fmt.Errorf("readlink: %w", err)
	}
	return target, nil
}
