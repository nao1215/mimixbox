// Package shred implements the shred applet: overwrite files repeatedly to make
// their previous contents hard to recover, optionally removing them afterwards.
package shred

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the shred applet.
type Command struct{}

// New returns a shred command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "shred" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Overwrite a file to hide its contents" }

// Run executes shred.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err)
	iterations := fs.IntP("iterations", "n", 3, "overwrite N times instead of the default 3")
	zero := fs.BoolP("zero", "z", false, "add a final overwrite with zeros to hide shredding")
	remove := fs.BoolP("remove", "u", false, "truncate and remove the file after overwriting")
	verbose := fs.BoolP("verbose", "v", false, "show progress")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *iterations < 0 {
		return command.Failuref("invalid number of passes: %d", *iterations)
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing file operand")
	}

	var firstErr error
	for _, name := range files {
		if err := c.shredFile(stdio, name, *iterations, *zero, *remove, *verbose); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "shred: %s: %v\n", name, err)
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// shredFile overwrites name with iterations random passes (plus an optional
// zero pass) and optionally removes it.
func (c *Command) shredFile(stdio command.IO, name string, iterations int, zero, remove, verbose bool) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory")
	}
	size := info.Size()

	f, err := os.OpenFile(name, os.O_WRONLY, 0) //nolint:gosec // overwriting a user-named file is the point
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	for pass := 0; pass < iterations; pass++ {
		if verbose {
			_, _ = fmt.Fprintf(stdio.Err, "shred: %s: pass %d/%d (random)\n", name, pass+1, iterations)
		}
		if err := overwrite(f, size, rand.Reader); err != nil {
			return err
		}
	}
	if zero {
		if verbose {
			_, _ = fmt.Fprintf(stdio.Err, "shred: %s: pass (zeros)\n", name)
		}
		if err := overwrite(f, size, zeroReader{}); err != nil {
			return err
		}
	}

	if remove {
		if err := f.Truncate(0); err != nil {
			return err
		}
		_ = f.Close()
		if err := os.Remove(name); err != nil {
			return err
		}
	}
	return nil
}

// overwrite rewinds f and writes size bytes drawn from src, syncing at the end.
func overwrite(f *os.File, size int64, src io.Reader) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.CopyN(f, src, size); err != nil && err != io.EOF {
		return err
	}
	return f.Sync()
}

// zeroReader is an infinite source of zero bytes used for the final zero pass.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
