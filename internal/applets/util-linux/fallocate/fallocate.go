// Package fallocate implements the fallocate applet: preallocate (or extend) the
// space for a file to a given length.
package fallocate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fallocate applet.
type Command struct{}

// New returns a fallocate command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fallocate" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Preallocate or extend space for a file" }

// Run executes fallocate.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-l LENGTH FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Ensure each FILE is at least LENGTH bytes, creating it if necessary. LENGTH " +
			"accepts a plain byte count or a K/M/G suffix (powers of 1024).",
		Examples: []command.Example{
			{Command: "fallocate -l 1M big.bin", Explain: "Make big.bin one mebibyte."},
		},
		ExitStatus: "0  success.\n1  the length was missing/invalid or a file could not be sized.",
	})
	length := fs.StringP("length", "l", "", "allocate LENGTH bytes (K/M/G suffixes allowed)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *length == "" {
		_, _ = fmt.Fprintln(stdio.Err, "fallocate: option -l (length) is required")
		return command.SilentFailure()
	}
	size, err := parseSize(*length)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "fallocate: %v\n", err)
		return command.SilentFailure()
	}

	files := fs.Args()
	if len(files) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "fallocate: missing file operand")
		return command.SilentFailure()
	}

	var failed bool
	for _, name := range files {
		if err := allocate(name, size); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "fallocate: %s\n", command.FileError(name, err))
			failed = true
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// allocate sizes name to at least size bytes, never shrinking it.
func allocate(name string, size int64) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() >= size {
		return nil // never shrink an existing file
	}
	return f.Truncate(size)
}

// parseSize parses a byte count with an optional K/M/G (1024-based) suffix.
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("invalid length %q", s)
	}
	mult := int64(1)
	switch s[len(s)-1] {
	case 'K', 'k':
		mult = 1024
	case 'M', 'm':
		mult = 1024 * 1024
	case 'G', 'g':
		mult = 1024 * 1024 * 1024
	}
	if mult != 1 {
		s = s[:len(s)-1]
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid length %q", s)
	}
	return n * mult, nil
}
