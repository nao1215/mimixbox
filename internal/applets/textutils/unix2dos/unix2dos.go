// Package unix2dos implements the unix2dos applet: convert the line endings of
// files from LF to CRLF, editing each file in place.
package unix2dos

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// Command is the unix2dos applet.
type Command struct{}

// New returns a unix2dos command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unix2dos" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change LF to CRLF" }

// Run executes unix2dos. Each operand must name a regular file; its LF line
// endings are rewritten to CRLF in place. A directory or missing file is
// reported on stderr and makes the command exit non-zero, but the remaining
// files are still processed.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var firstErr error
	for _, file := range fs.Args() {
		target := os.ExpandEnv(file)
		if !isRegularFile(target) {
			_, _ = fmt.Fprintln(stdio.Err, c.Name()+": skip "+target+": not regular file")
			firstErr = keep(firstErr)
			continue
		}

		lines, readErr := readFileToStrList(target)
		if readErr != nil {
			_, _ = fmt.Fprintln(stdio.Err, c.Name()+": "+target+": Can't read file and convert LF to CRLF")
			firstErr = keep(firstErr)
			continue
		}
		_, _ = fmt.Fprintln(stdio.Out, c.Name()+": converting file "+target+" to DOS format...")
		if writeErr := mb.ListToFile(target, toCRLF(lines)); writeErr != nil {
			_, _ = fmt.Fprintln(stdio.Err, writeErr)
			firstErr = keep(firstErr)
			continue
		}
	}
	return firstErr
}

// toCRLF replaces every LF that is not already part of a CRLF with CRLF.
func toCRLF(unixStr []string) []string {
	replaceStr := make([]string, 0, len(unixStr))
	for _, v := range unixStr {
		if strings.HasSuffix(v, "\r\n") {
			replaceStr = append(replaceStr, v)
		} else {
			replaceStr = append(replaceStr, strings.ReplaceAll(v, "\n", "\r\n"))
		}
	}
	return replaceStr
}

// isRegularFile reports whether path exists and is a regular file (not a
// directory).
func isRegularFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// readFileToStrList reads path and returns it split into lines, each line still
// carrying its trailing newline.
func readFileToStrList(path string) ([]string, error) {
	f, err := os.Open(path) //nolint:gosec // operating on a user-named file is the whole point
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var strList []string
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF && len(line) == 0 {
			break
		}
		strList = append(strList, line)
	}
	return strList, nil
}

// keep returns the first error seen, creating a silent failure when none exists
// yet (the user-facing message has already been printed).
func keep(existing error) error {
	if existing != nil {
		return existing
	}
	return command.SilentFailure()
}
