// Package pipeprogress implements the pipe_progress applet: copy standard input
// to standard output while printing progress dots to standard error, so a slow
// pipeline shows it is still making progress.
package pipeprogress

import (
	"context"
	"fmt"
	"io"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pipe_progress applet.
type Command struct{}

// New returns a pipe_progress command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pipe_progress" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Copy stdin to stdout, printing progress dots to stderr" }

// chunk is the read size; one dot is printed per chunk that carries data.
const chunk = 64 * 1024

// Run copies standard input to standard output, emitting a "." to standard error
// for each block read and a trailing newline at end of input.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Pass standard input through to standard output unchanged, printing a dot to " +
			"standard error for every block read. Useful as a heartbeat in a slow pipeline.",
		Examples: []command.Example{
			{Command: "tar c dir | pipe_progress | ssh host 'cat > backup.tar'", Explain: "Show progress while streaming."},
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	buf := make([]byte, chunk)
	wrote := false
	for {
		n, rerr := stdio.In.Read(buf)
		if n > 0 {
			if _, werr := stdio.Out.Write(buf[:n]); werr != nil {
				_, _ = fmt.Fprintf(stdio.Err, "pipe_progress: %v\n", werr)
				return command.SilentFailure()
			}
			_, _ = fmt.Fprint(stdio.Err, ".")
			wrote = true
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			_, _ = fmt.Fprintf(stdio.Err, "pipe_progress: %v\n", rerr)
			return command.SilentFailure()
		}
	}
	if wrote {
		_, _ = fmt.Fprintln(stdio.Err)
	}
	return nil
}
