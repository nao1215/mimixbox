// Package logcollect implements the log-collect applet: gather the log files
// present on the system into one output directory for inspection. It is a
// clean-room port of morrigan's log-collect.
package logcollect

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the log-collect applet.
type Command struct{}

// New returns a log-collect command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "log-collect" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Gather system log files into one directory" }

// Run executes log-collect.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [SOURCE]", stdio.Err)
	out := fs.StringP("output", "o", "collected-logs", "directory to copy the logs into")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	source := "/var/log"
	if rest := fs.Args(); len(rest) > 0 {
		source = rest[0]
	}

	copied, skipped, err := collect(source, *out)
	if err != nil {
		return command.Failuref("%v", err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "collected %d files into %s (%d skipped)\n", copied, *out, skipped)
	return nil
}

// collect walks source and copies every readable regular file into dst,
// recreating the directory layout. Unreadable files are skipped (root is
// usually required to read everything under /var/log). It returns the number of
// files copied and skipped.
func collect(source, dst string) (copied, skipped int, err error) {
	root, err := os.Stat(source)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read source %q: %v", source, err)
	}
	if !root.IsDir() {
		return 0, 0, fmt.Errorf("source %q is not a directory", source)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil { //nolint:gosec // a collection directory is not sensitive
		return 0, 0, fmt.Errorf("cannot create %q: %v", dst, err)
	}

	walkErr := filepath.Walk(source, func(path string, fi os.FileInfo, werr error) error {
		if werr != nil || fi.IsDir() || !fi.Mode().IsRegular() {
			if werr != nil {
				skipped++
			}
			return nil //nolint:nilerr // keep walking past unreadable entries
		}
		rel, rerr := filepath.Rel(source, path)
		if rerr != nil {
			skipped++
			return nil
		}
		if copyFile(path, filepath.Join(dst, rel)) != nil {
			skipped++
			return nil
		}
		copied++
		return nil
	})
	return copied, skipped, walkErr
}

// copyFile copies src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // reading a discovered log file is the point
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { //nolint:gosec // mirror layout
		return err
	}
	out, err := os.Create(dst) //nolint:gosec // writing into the collection directory
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
