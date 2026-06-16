// Package truncate implements the truncate applet: shrink or extend each file
// to a given size, creating it when necessary.
package truncate

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the truncate applet.
type Command struct{}

// New returns a truncate command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "truncate" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Shrink or extend the size of a file to a given size" }

// Run executes truncate.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-s SIZE FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Shrink or extend the size of each FILE to SIZE. SIZE may carry a K, M, or G suffix; " +
			"a leading + or - adjusts the size relative to its current value. A missing FILE is created, unless -c is given.",
		Examples: []command.Example{
			{Command: "truncate -s 1K data.bin", Explain: "Set the file size to exactly 1024 bytes."},
			{Command: "truncate -s +10M log.txt", Explain: "Grow the file by 10 MiB."},
			{Command: "truncate -c -s 0 existing.log", Explain: "Empty the file but do not create it if absent."},
		},
		ExitStatus: "0  all files were resized successfully.\n1  a size was invalid or a file could not be resized.",
	})
	sizeSpec := fs.StringP("size", "s", "", "set or adjust the file size by SIZE bytes")
	noCreate := fs.BoolP("no-create", "c", false, "do not create any files")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *sizeSpec == "" {
		return command.Failuref("you must specify a size with -s")
	}

	files := fs.Args()
	if len(files) == 0 {
		return command.Failuref("missing file operand")
	}

	var firstErr error
	for _, name := range files {
		if err := c.truncateFile(name, *sizeSpec, *noCreate); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "truncate: %v\n", err)
			if firstErr == nil {
				firstErr = command.SilentFailure()
			}
		}
	}
	return firstErr
}

// truncateFile resizes name according to spec. A leading '+' or '-' adjusts the
// current size; otherwise spec is the absolute target size.
func (c *Command) truncateFile(name, spec string, noCreate bool) error {
	info, statErr := os.Stat(name)
	if os.IsNotExist(statErr) {
		if noCreate {
			return nil
		}
	} else if statErr != nil {
		return fmt.Errorf("cannot stat %q: %v", name, statErr)
	}

	var cur int64
	if info != nil {
		cur = info.Size()
	}
	target, err := resolveSize(spec, cur)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0o644) //nolint:gosec // creating a user-named file is the point
	if err != nil {
		return fmt.Errorf("cannot open %q: %v", name, err)
	}
	defer func() { _ = f.Close() }()

	if err := f.Truncate(target); err != nil {
		return fmt.Errorf("cannot truncate %q: %v", name, err)
	}
	return nil
}

// resolveSize turns a size spec into an absolute target, given the file's
// current size. It understands a leading +/- and the K/M/G/KB/MB/GB suffixes.
func resolveSize(spec string, cur int64) (int64, error) {
	rel := 0
	switch spec[0] {
	case '+':
		rel, spec = 1, spec[1:]
	case '-':
		rel, spec = -1, spec[1:]
	}

	n, err := parseBytes(spec)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", spec)
	}

	var target int64
	switch rel {
	case 1:
		target = cur + n
	case -1:
		target = cur - n
	default:
		target = n
	}
	if target < 0 {
		target = 0
	}
	return target, nil
}

// parseBytes parses a non-negative byte count with an optional binary suffix.
func parseBytes(s string) (int64, error) {
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "K"), strings.HasSuffix(s, "KB"):
		mult, s = 1024, strings.TrimRight(s, "KB")
	case strings.HasSuffix(s, "M"), strings.HasSuffix(s, "MB"):
		mult, s = 1024*1024, strings.TrimRight(s, "MB")
	case strings.HasSuffix(s, "G"), strings.HasSuffix(s, "GB"):
		mult, s = 1024*1024*1024, strings.TrimRight(s, "GB")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid number")
	}
	return n * mult, nil
}
