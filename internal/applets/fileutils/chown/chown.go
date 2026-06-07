// Package chown implements the chown applet: change the owner and/or group of
// each FILE to OWNER and/or GROUP, with the common GNU options.
package chown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chown applet.
type Command struct{}

// New returns a chown command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chown" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Change the owner and/or group of each FILE to OWNER and/or GROUP"
}

// Run executes chown.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... OWNER[:GROUP] FILE...", stdio.Err)
	recursive := fs.BoolP("recursive", "R", false, "operate on files and directories recursively")
	verbose := fs.BoolP("verbose", "v", false, "output a diagnostic for every file processed")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand\n", c.Name())
		return command.SilentFailure()
	}

	spec := operands[0]
	files := operands[1:]
	if len(files) == 0 {
		_, _ = fmt.Fprintf(stdio.Err, "%s: missing operand after '%s'\n", c.Name(), spec)
		return command.SilentFailure()
	}

	uid, gid, err := parseOwner(spec)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	var firstErr error
	for _, path := range files {
		if err := apply(stdio, path, uid, gid, *recursive, *verbose); err != nil {
			firstErr = keep(firstErr)
		}
	}
	return firstErr
}

// apply changes the ownership of a single operand, optionally recursing into
// directories. A failure is reported on stderr GNU-style and returns an error
// that only sets the exit code, so the caller keeps processing later files.
func apply(stdio command.IO, path string, uid, gid int, recursive, verbose bool) error {
	if recursive {
		return filepath.Walk(path, func(p string, _ os.FileInfo, err error) error {
			if err != nil {
				_, _ = fmt.Fprintf(stdio.Err, "chown: %s\n", command.FileError(p, err))
				return command.SilentFailure()
			}
			return chownOne(stdio, p, uid, gid, verbose)
		})
	}
	return chownOne(stdio, path, uid, gid, verbose)
}

// chownOne performs the os.Chown for a single path and renders the GNU-style
// diagnostics for permission failures and verbose mode.
func chownOne(stdio command.IO, path string, uid, gid int, verbose bool) error {
	if err := os.Chown(path, uid, gid); err != nil {
		if errors.Is(err, os.ErrPermission) {
			_, _ = fmt.Fprintf(stdio.Err, "chown: changing ownership of '%s': Operation not permitted\n", path)
		} else {
			_, _ = fmt.Fprintf(stdio.Err, "chown: %s\n", command.FileError(path, err))
		}
		return command.SilentFailure()
	}
	if verbose {
		_, _ = fmt.Fprintf(stdio.Out, "ownership of '%s' retained\n", path)
	}
	return nil
}

// parseOwner parses a GNU chown spec of the form OWNER, OWNER:GROUP, OWNER:, or
// :GROUP, where each part is a name or a numeric id. It returns the resolved
// uid and gid, using -1 for any part left unspecified (so os.Chown leaves it
// unchanged). Resolution goes through os/user so it works without root.
func parseOwner(spec string) (uid, gid int, err error) {
	uid, gid = -1, -1
	if spec == "" {
		return -1, -1, errors.New("missing operand")
	}

	ownerPart := spec
	var groupPart string
	hasGroup := false
	if i := strings.IndexByte(spec, ':'); i >= 0 {
		ownerPart = spec[:i]
		groupPart = spec[i+1:]
		hasGroup = true
	}

	if ownerPart != "" {
		uid, err = lookupUID(ownerPart)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid user: '%s'", spec)
		}
	}

	if hasGroup && groupPart != "" {
		gid, err = lookupGID(groupPart)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid group: '%s'", spec)
		}
	}

	return uid, gid, nil
}

// lookupUID resolves a user name to a uid, accepting a numeric id directly.
func lookupUID(name string) (int, error) {
	if u, err := user.Lookup(name); err == nil {
		return strconv.Atoi(u.Uid)
	}
	if id, err := strconv.Atoi(name); err == nil {
		return id, nil
	}
	return -1, fmt.Errorf("unknown user: %s", name)
}

// lookupGID resolves a group name to a gid, accepting a numeric id directly.
func lookupGID(name string) (int, error) {
	if g, err := user.LookupGroup(name); err == nil {
		return strconv.Atoi(g.Gid)
	}
	if id, err := strconv.Atoi(name); err == nil {
		return id, nil
	}
	return -1, fmt.Errorf("unknown group: %s", name)
}

func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
