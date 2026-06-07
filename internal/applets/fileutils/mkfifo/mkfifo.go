// Package mkfifo implements the mkfifo applet: create FIFOs (named pipes) at the
// given paths, with the common GNU options.
package mkfifo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// defaultMode is the mode FIFOs are created with when -m/--mode is not given.
// 0644 yields prw-r--r-- once the umask is applied by the system.
const defaultMode = 0o644

// Command is the mkfifo applet.
type Command struct{}

// New returns a mkfifo command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "mkfifo" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Make FIFO (named pipe)" }

// Run executes mkfifo.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... NAME...", stdio.Err)
	modeStr := fs.StringP("mode", "m", "", "set file permission bits to MODE, not a=rw - umask")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		fmt.Fprintf(stdio.Err, "%s: missing operand\n", c.Name())
		return command.SilentFailure()
	}

	mode := os.FileMode(defaultMode)
	if *modeStr != "" {
		m, perr := parseMode(*modeStr)
		if perr != nil {
			fmt.Fprintf(stdio.Err, "%s: invalid mode: %q\n", c.Name(), *modeStr)
			return command.SilentFailure()
		}
		mode = m
	}

	var failErr error
	for _, name := range names {
		path := os.ExpandEnv(name)
		if err := makeFifo(path, mode); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %s\n", c.Name(), fifoError(path, err))
			failErr = command.SilentFailure()
			continue
		}
	}
	return failErr
}

// makeFifo creates a single FIFO at path. It reports an "already exist" sentinel
// when the path is already present so the caller can render the GNU-style
// message the integration spec asserts.
func makeFifo(path string, mode os.FileMode) error {
	if _, err := os.Lstat(path); err == nil {
		return errExist
	}
	return syscall.Mkfifo(path, uint32(mode))
}

// errExist marks an attempt to create a FIFO whose path already exists.
var errExist = errors.New("already exist")

// fifoError formats a failed FIFO creation the way the integration spec expects:
// an existing path becomes "<path>: already exist", and any other failure becomes
// "<path>: <reason>" with the underlying syscall message lower-cased.
func fifoError(path string, err error) string {
	if errors.Is(err, errExist) {
		return "can't make " + path + ": already exist"
	}
	return path + ": " + reason(err)
}

// reason extracts the human-readable message from a syscall error, matching the
// lower-case GNU style ("no such file or directory").
func reason(err error) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno.Error()
	}
	return err.Error()
}

// parseMode parses an octal MODE string such as "644" or "0755".
func parseMode(s string) (os.FileMode, error) {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(v), nil
}
