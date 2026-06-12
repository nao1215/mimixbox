// Package svlogd implements the svlogd applet: read standard input and append it
// to a log directory's current file, rotating it by size.
package svlogd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the svlogd applet.
type Command struct{}

// New returns a svlogd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "svlogd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Log standard input to a directory" }

// Injected so the rotation size and the rotated-file timestamp are testable.
var (
	maxSize int64 = 1000000 // rotate the current file when it reaches this many bytes
	now           = time.Now
)

// Run executes svlogd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t] DIR", stdio.Err).WithHelp(command.Help{
		Description: "Read standard input line by line and append each line to DIR/current, rotating " +
			"that file to a timestamped name once it grows past the size limit. With -t each line is " +
			"prefixed with the current time. This is a single-directory subset of the runit svlogd.",
		Examples: []command.Example{
			{Command: "mydaemon | svlogd /var/log/mydaemon", Explain: "Log a program's output."},
		},
		ExitStatus: "0  standard input reached EOF.\n1  no directory was given or the log was unwritable.",
	})
	timestamp := fs.BoolP("timestamp", "t", false, "prefix each line with a timestamp")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a log directory is required")
	}
	dir := rest[0]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return command.Failuref("cannot create %s: %v", dir, err)
	}

	w, err := newWriter(dir)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer func() { _ = w.close() }()

	sc := bufio.NewScanner(stdio.In)
	for sc.Scan() {
		line := sc.Text()
		if *timestamp {
			line = now().UTC().Format("2006-01-02_15:04:05.000000000") + " " + line
		}
		if err := w.write(line + "\n"); err != nil {
			return command.Failuref("cannot write the log: %v", err)
		}
	}
	return sc.Err()
}

// writer appends to DIR/current and rotates it by size.
type writer struct {
	dir     string
	current string
	f       *os.File
	size    int64
}

func newWriter(dir string) (*writer, error) {
	current := filepath.Join(dir, "current")
	f, err := os.OpenFile(current, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // log file
	if err != nil {
		return nil, err
	}
	info, _ := f.Stat()
	return &writer{dir: dir, current: current, f: f, size: sizeOf(info)}, nil
}

func (w *writer) write(s string) error {
	if w.size+int64(len(s)) > maxSize && w.size > 0 {
		if err := w.rotate(); err != nil {
			return err
		}
	}
	n, err := w.f.WriteString(s)
	w.size += int64(n)
	return err
}

// rotate closes the current file, renames it to a timestamped name, and opens a
// fresh current file.
func (w *writer) rotate() error {
	if err := w.f.Close(); err != nil {
		return err
	}
	rotated := filepath.Join(w.dir, fmt.Sprintf("@%d.s", now().UnixNano()))
	if err := os.Rename(w.current, rotated); err != nil {
		return err
	}
	f, err := os.OpenFile(w.current, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // log file
	if err != nil {
		return err
	}
	w.f = f
	w.size = 0
	return nil
}

func (w *writer) close() error { return w.f.Close() }

func sizeOf(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}
