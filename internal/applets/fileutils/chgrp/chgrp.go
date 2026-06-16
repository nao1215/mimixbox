// Package chgrp implements the chgrp applet: change the group ownership of each
// FILE to GROUP, with the common GNU options.
package chgrp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chgrp applet.
type Command struct{}

// New returns a chgrp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chgrp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change the group of each FILE to GROUP" }

type options struct {
	recursive bool
	verbose   bool
}

// Run executes chgrp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... GROUP FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Change the group ownership of each FILE to GROUP. GROUP may be a group name or a numeric group ID.",
		Examples: []command.Example{
			{Command: "chgrp staff report.txt", Explain: "Change the group of report.txt to staff."},
			{Command: "chgrp -R wheel /srv/www", Explain: "Recursively change the group of /srv/www and its contents."},
			{Command: "chgrp -v 1000 file", Explain: "Change the group to GID 1000, reporting the change."},
		},
		ExitStatus: "0  all files were changed successfully.\n1  one or more files could not be changed.",
	})
	recursive := fs.BoolP("recursive", "R", false, "operate on files and directories recursively")
	verbose := fs.BoolP("verbose", "v", false, "output a diagnostic for every file processed")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand\n", c.Name())
		return command.SilentFailure()
	}

	opts := options{recursive: *recursive, verbose: *verbose}

	group := rest[0]
	gid, ok := lookupGid(group)
	if !ok {
		_, _ = fmt.Fprintf(stdio.Err, "%s: invalid group: '%s'\n", c.Name(), group)
		return command.SilentFailure()
	}

	return c.chgrp(stdio, gid, rest[1:], opts)
}

// lookupGid resolves a group name or a numeric gid to its integer gid. The
// second return value reports whether the group could be resolved.
func lookupGid(group string) (int, bool) {
	if g, err := user.LookupGroup(group); err == nil {
		if gid, err := strconv.Atoi(g.Gid); err == nil {
			return gid, true
		}
	}
	// Fall back to a numeric gid even if it has no /etc/group entry.
	if gid, err := strconv.Atoi(group); err == nil {
		return gid, true
	}
	return 0, false
}

// chgrp changes the group of every file, continuing past any failures. A failed
// file is reported GNU-style on stderr; the returned error only sets the exit
// code, because its message was already printed.
func (c *Command) chgrp(stdio command.IO, gid int, files []string, opts options) error {
	var failErr error
	for _, path := range files {
		path = os.ExpandEnv(path)
		var err error
		if opts.recursive {
			err = c.changeGroupRecursive(stdio, path, gid, opts)
		} else {
			err = c.changeGroup(stdio, path, gid, opts)
		}
		if err != nil {
			failErr = command.SilentFailure()
		}
	}
	return failErr
}

func (c *Command) changeGroupRecursive(stdio command.IO, path string, gid int, opts options) error {
	var walkErr error
	err := filepath.Walk(path, func(p string, _ os.FileInfo, err error) error {
		if err != nil {
			c.report(stdio, p, err)
			walkErr = err
			return nil
		}
		if cerr := c.changeGroup(stdio, p, gid, opts); cerr != nil {
			walkErr = cerr
		}
		return nil
	})
	if err != nil {
		return err
	}
	return walkErr
}

func (c *Command) changeGroup(stdio command.IO, path string, gid int, opts options) error {
	// uid -1 leaves the owner unchanged.
	if err := os.Chown(path, -1, gid); err != nil {
		c.report(stdio, path, err)
		return err
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(stdio.Out, "changed group of '%s'\n", path)
	}
	return nil
}

// report writes a GNU-style diagnostic for a failed chown to stderr.
func (c *Command) report(stdio command.IO, path string, err error) {
	var pe *os.PathError
	if errors.As(err, &pe) {
		err = pe.Err
	}
	if errors.Is(err, os.ErrPermission) {
		_, _ = fmt.Fprintf(stdio.Err, "%s: changing group of '%s': Operation not permitted\n", c.Name(), path)
		return
	}
	_, _ = fmt.Fprintf(stdio.Err, "%s: changing group of '%s': %v\n", c.Name(), path, err)
}
