// Package dos2unix implements the dos2unix applet: convert CRLF line endings to
// LF, editing each named file in place.
package dos2unix

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	mb "github.com/nao1215/mimixbox/internal/lib"
)

// Command is the dos2unix applet.
type Command struct{}

// New returns a dos2unix command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dos2unix" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change CRLF to LF" }

// Run executes dos2unix. Each operand is a regular file whose CRLF line endings
// are rewritten to LF in place. A directory or missing file is reported on
// stderr and makes the command exit non-zero; the remaining files are still
// converted.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... FILE...", stdio.Err).WithHelp(command.Help{
		Description: "Convert CRLF (DOS) line endings to LF (Unix), editing each named FILE in place.",
		Examples: []command.Example{
			{Command: "dos2unix file.txt", Explain: "Convert file.txt to Unix line endings in place."},
			{Command: "dos2unix a.txt b.txt", Explain: "Convert several files in place."},
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. a file was missing or could not be written).",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	return convert(stdio, fs.Args())
}

// convert rewrites CRLF to LF for each named file in place. It mirrors the
// historical behavior and messages the integration tests depend on: a skipped
// (non-regular) file or a read/write failure is reported on stderr but does not
// stop the remaining files, and any failure sets the exit code to 1.
func convert(stdio command.IO, files []string) error {
	failed := false
	for _, file := range files {
		target := os.ExpandEnv(file)
		if !mb.IsFile(target) {
			_, _ = fmt.Fprintln(stdio.Err, "dos2unix: skip "+target+": not regular file")
			failed = true
			continue
		}

		lines, err := mb.ReadFileToStrList(target)
		if err != nil {
			_, _ = fmt.Fprintln(stdio.Err, "dos2unix: "+target+": can't read file and convert CRLF to LF")
			failed = true
			continue
		}
		_, _ = fmt.Fprintln(stdio.Out, "dos2unix: converting file "+target+" to Unix format...")
		lines = toLF(lines)
		if err := mb.ListToFile(target, lines); err != nil {
			_, _ = fmt.Fprintln(stdio.Err, err)
			failed = true
			continue
		}
	}

	if failed {
		return command.SilentFailure()
	}
	return nil
}

// toLF replaces a trailing CRLF on each line with LF.
func toLF(dosStr []string) []string {
	replaceStr := make([]string, 0, len(dosStr))
	for _, v := range dosStr {
		if strings.HasSuffix(v, "\r\n") {
			replaceStr = append(replaceStr, strings.ReplaceAll(v, "\r\n", "\n"))
		} else {
			replaceStr = append(replaceStr, v)
		}
	}
	return replaceStr
}
