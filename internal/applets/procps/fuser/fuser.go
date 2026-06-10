// Package fuser implements the fuser applet: identify the processes using a file
// (or directory) by scanning their /proc entries.
package fuser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fuser applet.
type Command struct{}

// New returns a fuser command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fuser" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Identify processes using a file" }

// procDir is the /proc mount; tests point it at a fixture.
var procDir = "/proc"

// Run executes fuser.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Print the IDs of the processes that have FILE open, or use it as their working " +
			"directory, executable, or root, found by scanning /proc. The 'FILE:' header is written " +
			"to standard error and the PIDs to standard output.",
		Examples: []command.Example{
			{Command: "fuser /var/log/syslog", Explain: "List the processes using the log file."},
		},
		ExitStatus: "0  at least one process was found for some FILE.\n1  none were found.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "fuser: a FILE is required")
		return command.SilentFailure()
	}

	anyFound := false
	for _, name := range files {
		target := resolve(name)
		pids := usersOf(target)
		_, _ = fmt.Fprintf(stdio.Err, "%s:", name)
		var parts []string
		for _, pid := range pids {
			parts = append(parts, strconv.Itoa(pid))
		}
		if len(parts) > 0 {
			anyFound = true
			_, _ = fmt.Fprintln(stdio.Out, strings.Join(parts, " "))
		}
	}

	if !anyFound {
		return command.SilentFailure()
	}
	return nil
}

// resolve returns the canonical absolute path of name for comparison.
func resolve(name string) string {
	if abs, err := filepath.Abs(name); err == nil {
		name = abs
	}
	if real, err := filepath.EvalSymlinks(name); err == nil {
		return real
	}
	return name
}

// usersOf returns the sorted PIDs whose fd/cwd/exe/root point at target.
func usersOf(target string) []int {
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return nil
	}
	var pids []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if processUses(filepath.Join(procDir, e.Name()), target) {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	return pids
}

// processUses reports whether the process at procPath references target through
// its cwd, exe, root, or any open file descriptor.
func processUses(procPath, target string) bool {
	for _, link := range []string{"cwd", "exe", "root"} {
		if linkEquals(filepath.Join(procPath, link), target) {
			return true
		}
	}
	fds, err := os.ReadDir(filepath.Join(procPath, "fd"))
	if err != nil {
		return false
	}
	for _, fd := range fds {
		if linkEquals(filepath.Join(procPath, "fd", fd.Name()), target) {
			return true
		}
	}
	return false
}

// linkEquals reports whether the symlink at path resolves to target.
func linkEquals(path, target string) bool {
	dest, err := os.Readlink(path)
	if err != nil {
		return false
	}
	if real, err := filepath.EvalSymlinks(dest); err == nil {
		dest = real
	}
	return dest == target
}
