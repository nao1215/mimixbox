// Package yes implements the yes applet: repeatedly output a line until it is
// killed, printing "y" by default or the joined command-line arguments.
package yes

import (
	"bufio"
	"context"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the yes applet.
type Command struct{}

// New returns a yes command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "yes" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Output a string repeatedly until killed" }

// Run executes yes. It writes the line forever and returns only when the output
// stream reports an error (for example a closed pipe), so it terminates the way
// the system yes does when its reader goes away.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[STRING]...", stdio.Err).WithHelp(command.Help{
		Description: "Repeatedly print a line until killed, writing \"y\" by default or the STRING " +
			"operands joined by spaces. Useful for feeding an affirmative answer to an interactive program.",
		Examples: []command.Example{
			{Command: "yes", Explain: "Print \"y\" repeatedly until interrupted."},
			{Command: "yes no", Explain: "Print \"no\" repeatedly until interrupted."},
		},
		ExitStatus: "0  output ended normally (the reader closed the pipe).\n1  an error occurred.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	line := "y"
	if rest := fs.Args(); len(rest) > 0 {
		line = strings.Join(rest, " ")
	}
	line += "\n"

	w := bufio.NewWriter(stdio.Out)
	for {
		select {
		case <-ctx.Done():
			_ = w.Flush()
			return nil
		default:
		}
		if _, err := w.WriteString(line); err != nil {
			return nil
		}
	}
}
