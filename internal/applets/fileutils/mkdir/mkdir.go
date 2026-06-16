// Package mkdir implements the mkdir applet: create directories, optionally
// creating parent directories as needed.
package mkdir

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mkdir applet.
type Command struct{}

// New returns a mkdir command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkdir" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Make directories" }

// Run executes mkdir.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... DIRECTORY...", stdio.Err).WithHelp(command.Help{
		Description: "Create each named DIRECTORY if it does not already exist. With -p, create any missing parent directories and do not error when a directory already exists.",
		Examples: []command.Example{
			{Command: "mkdir -p a/b/c", Explain: "Create a/b/c, making parents as needed."},
			{Command: "mkdir -m 700 secret", Explain: "Create 'secret' with mode 0700."},
		},
		ExitStatus: "0  every directory was created.\n1  a directory could not be created.",
	})
	parents := fs.BoolP("parents", "p", false, "no error if existing, make parent directories as needed")
	verbose := fs.BoolP("verbose", "v", false, "print a message for each created directory")
	mode := fs.StringP("mode", "m", "", "set file mode (as in chmod), not a=rwx - umask")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "mkdir: no operand")
		return command.SilentFailure()
	}

	perm := os.FileMode(0o755)
	if *mode != "" {
		parsed, perr := strconv.ParseUint(*mode, 8, 32)
		if perr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "mkdir: invalid mode '%s'\n", *mode)
			return command.SilentFailure()
		}
		perm = os.FileMode(parsed)
	}

	var firstErr error
	for _, path := range operands {
		target := os.ExpandEnv(path)
		var mkErr error
		if *parents {
			mkErr = os.MkdirAll(target, perm)
		} else {
			mkErr = os.Mkdir(target, perm)
		}
		if mkErr != nil {
			_, _ = fmt.Fprintln(stdio.Err, mkErr.Error())
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
			continue
		}
		if *verbose {
			_, _ = fmt.Fprintf(stdio.Out, "mkdir: created directory '%s'\n", target)
		}
	}
	return firstErr
}
