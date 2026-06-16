// Package which implements the which applet: locate a command on $PATH and
// print the absolute path that would be executed.
package which

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the which applet.
type Command struct{}

// New returns a which command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "which" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Returns the file path which would be executed in the current environment"
}

// Run executes which: for each COMMAND operand it looks the name up on $PATH and
// prints the path that would be executed, one per line. With -a/--all every
// match across PATH is printed. If any operand is not found (or none is given),
// the exit status is 1, matching the Debian which behavior: nothing is written
// for a name that cannot be resolved.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... COMMAND...", stdio.Err).WithHelp(command.Help{
		Description: "For each COMMAND, print the absolute path that would be executed for it on the " +
			"current $PATH. With -a, print every matching path instead of only the first.",
		Examples: []command.Example{
			{Command: "which ls", Explain: "Print the path of the ls command."},
			{Command: "which -a python", Explain: "Print every python found on $PATH."},
		},
		ExitStatus: "0  every COMMAND was found.\n1  a COMMAND was not found, or none was given.",
	})
	all := fs.BoolP("all", "a", false, "print all matching pathnames of each argument")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		// No operand: nothing to print and a non-zero exit status.
		return command.SilentFailure()
	}

	found := true
	for _, name := range names {
		if *all {
			matches := lookPathAll(name)
			if len(matches) == 0 {
				found = false
				continue
			}
			for _, p := range matches {
				_, _ = fmt.Fprintln(stdio.Out, p)
			}
			continue
		}
		p, lerr := exec.LookPath(name)
		if lerr != nil {
			// Don't print anything for a name that is not found.
			found = false
			continue
		}
		_, _ = fmt.Fprintln(stdio.Out, p)
	}

	if !found {
		return command.SilentFailure()
	}
	return nil
}

// lookPathAll returns every executable named name found across the directories
// in $PATH, in PATH order. A name containing a path separator is resolved
// directly without consulting $PATH.
func lookPathAll(name string) []string {
	if strings.ContainsRune(name, os.PathSeparator) {
		if isExecutable(name) {
			if abs, err := filepath.Abs(name); err == nil {
				return []string{abs}
			}
			return []string{name}
		}
		return nil
	}

	var matches []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, name)
		if isExecutable(candidate) {
			matches = append(matches, candidate)
		}
	}
	return matches
}

// isExecutable reports whether path is a regular file with an executable bit set.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return mode.IsRegular() && mode.Perm()&0o111 != 0
}
