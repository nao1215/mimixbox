// Package pwd implements the pwd applet: print the name of the current working
// directory.
package pwd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pwd applet.
type Command struct{}

// New returns a pwd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pwd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print Working Directory" }

// Run executes pwd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)
	logical := fs.BoolP("logical", "L", false, "use PWD from environment, even if it contains symlinks")
	physical := fs.BoolP("physical", "P", false, "avoid all symlinks")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	dir, err := workingDir(*physical && !*logical)
	if err != nil {
		fmt.Fprintf(stdio.Err, "pwd: %v\n", err)
		return command.SilentFailure()
	}
	fmt.Fprintln(stdio.Out, dir)
	return nil
}

// workingDir returns the current working directory. By default (logical) it
// honours $PWD when it names the current directory; with physical true it
// resolves all symbolic links to an absolute canonical path.
func workingDir(physical bool) (string, error) {
	if physical {
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(dir)
	}

	if pwd := os.Getenv("PWD"); filepath.IsAbs(pwd) && namesCurrentDir(pwd) {
		return pwd, nil
	}
	return os.Getwd()
}

// namesCurrentDir reports whether pwd refers to the current working directory,
// i.e. resolving its symlinks yields the same canonical directory as the
// process's actual working directory.
func namesCurrentDir(pwd string) bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}
	rp, err := filepath.EvalSymlinks(pwd)
	if err != nil {
		return false
	}
	rwd, err := filepath.EvalSymlinks(wd)
	if err != nil {
		return false
	}
	return rp == rwd
}
