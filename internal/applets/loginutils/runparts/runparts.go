// Package runparts implements the run-parts applet: run every valid executable
// in a directory, in alphabetical order.
package runparts

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the run-parts applet.
type Command struct{}

// New returns a run-parts command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "run-parts" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run all executables in a directory" }

// validName matches the names run-parts accepts (LSB rule): letters, digits,
// underscores, and hyphens only — so editor backups and packaging leftovers
// (foo~, foo.bak, foo.dpkg-new) are skipped.
var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Run executes run-parts.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[--test] [--list] [--arg=ARG]... DIRECTORY", stdio.Err).WithHelp(command.Help{
		Description: "Run every executable in DIRECTORY whose name contains only letters, digits, " +
			"underscores, and hyphens, in alphabetical order. --list prints the matching files " +
			"without running them; --test prints those that would run (matching and executable); " +
			"--arg adds an argument passed to each program; --verbose announces each one before " +
			"running it.",
		Examples: []command.Example{
			{Command: "run-parts /etc/cron.daily", Explain: "Run the daily cron scripts in order."},
			{Command: "run-parts --test /etc/cron.daily", Explain: "Show what would run."},
		},
		ExitStatus: "0  all programs succeeded.\n1  no directory was given or a program failed.",
	})
	test := fs.Bool("test", false, "print the programs that would run, without running them")
	list := fs.Bool("list", false, "list the matching files without running them")
	verbose := fs.BoolP("verbose", "v", false, "print each program name before running it")
	extraArgs := fs.StringArray("arg", nil, "argument to pass to each program (repeatable)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a directory is required")
	}
	dir := rest[0]

	names, err := matching(dir)
	if err != nil {
		return command.Failuref("%s: %v", dir, err)
	}

	failed := false
	for _, name := range names {
		path := filepath.Join(dir, name)
		executable := isExecutable(path)

		if *list {
			_, _ = fmt.Fprintln(stdio.Out, path)
			continue
		}
		if !executable {
			continue
		}
		if *test {
			_, _ = fmt.Fprintln(stdio.Out, path)
			continue
		}
		if *verbose {
			_, _ = fmt.Fprintf(stdio.Err, "run-parts: executing %s\n", path)
		}
		cmd := exec.Command(path, *extraArgs...) //nolint:gosec // running the directory's programs is the point
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		if err := cmd.Run(); err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				_, _ = fmt.Fprintf(stdio.Err, "run-parts: %s exited with code %d\n", path, ee.ExitCode())
			} else {
				_, _ = fmt.Fprintf(stdio.Err, "run-parts: %s: %v\n", path, err)
			}
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// matching returns the valid run-parts file names in dir, sorted.
func matching(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !validName.MatchString(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}
