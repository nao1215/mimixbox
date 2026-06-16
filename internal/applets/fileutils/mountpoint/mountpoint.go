// Package mountpoint implements the mountpoint applet: report whether a given
// directory is the mount point of a file system.
package mountpoint

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the mountpoint applet.
type Command struct{}

// New returns a mountpoint command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mountpoint" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "See if a directory is a mountpoint" }

// Run executes mountpoint.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... DIRECTORY", stdio.Err).WithHelp(command.Help{
		Description: "Report whether DIRECTORY is the mount point of a file system. The exit status reflects the answer, so the command is useful in shell tests; with -q it prints nothing and only sets the exit code.",
		Examples: []command.Example{
			{Command: "mountpoint /mnt", Explain: "Print whether /mnt is a mountpoint."},
			{Command: "mountpoint -q /mnt && echo mounted", Explain: "Run a command only when /mnt is a mountpoint."},
		},
		ExitStatus: "0  DIRECTORY is a mountpoint.\n1  DIRECTORY is not a mountpoint, or an error occurred.",
	})
	quiet := fs.BoolP("quiet", "q", false, "be quiet, only set the exit code")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	dirs := fs.Args()
	if len(dirs) != 1 {
		return command.Failuref("exactly one argument is required")
	}
	dir := dirs[0]

	isMount, err := mounted(dir)
	if err != nil {
		if *quiet {
			return command.SilentFailure()
		}
		return command.Failuref("%v", err)
	}

	if !*quiet {
		if isMount {
			_, _ = fmt.Fprintf(stdio.Out, "%s is a mountpoint\n", dir)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "%s is not a mountpoint\n", dir)
		}
	}
	if !isMount {
		return &command.ExitError{Code: command.ExitFailure}
	}
	return nil
}

// mounted reports whether dir is a mount point. A directory is a mount point
// when it sits on a different device than its parent, or when it is its own
// parent (the file-system root).
func mounted(dir string) (bool, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		return false, fmt.Errorf("%s: %v", dir, err)
	}
	var parent syscall.Stat_t
	if err := syscall.Stat(filepath.Dir(dir), &parent); err != nil {
		return false, fmt.Errorf("%s: %v", filepath.Dir(dir), err)
	}
	if st.Dev != parent.Dev {
		return true, nil
	}
	// Same device but identical inode means dir is its own parent: the root.
	return st.Ino == parent.Ino, nil
}
