// Package ts implements the ts applet: prefix each line of standard input with a
// timestamp.
package ts

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ts applet.
type Command struct{}

// New returns a ts command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ts" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Timestamp each input line" }

// now is indirected so the timestamps are deterministic in tests.
var now = time.Now

const defaultLayout = "2006-01-02 15:04:05"

// Run executes ts.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-r]", stdio.Err).WithHelp(command.Help{
		Description: "Read standard input line by line and print each line prefixed with the current " +
			"time (YYYY-MM-DD HH:MM:SS). With -r the prefix is instead the time elapsed since the " +
			"first line, as a duration. This is the moreutils ts.",
		Examples: []command.Example{
			{Command: "tail -f log | ts", Explain: "Timestamp a log stream as it arrives."},
		},
		ExitStatus: "0  standard input reached EOF.",
	})
	relative := fs.BoolP("relative", "r", false, "show the time elapsed since the first line")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var start time.Time
	sc := bufio.NewScanner(stdio.In)
	for sc.Scan() {
		prefix := now().Format(defaultLayout)
		if *relative {
			if start.IsZero() {
				start = now()
			}
			prefix = now().Sub(start).String()
		}
		if _, err := fmt.Fprintf(stdio.Out, "%s %s\n", prefix, sc.Text()); err != nil {
			return command.Failuref("%v", err)
		}
	}
	return sc.Err()
}
