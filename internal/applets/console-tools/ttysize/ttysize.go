// Package ttysize implements the ttysize applet: print the terminal's width and
// height in character cells.
package ttysize

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the ttysize applet.
type Command struct{}

// New returns a ttysize command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ttysize" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the terminal width and height" }

// Default size used when the terminal size cannot be determined.
const (
	defaultWidth  = 80
	defaultHeight = 24
)

// getSizeFn is indirected so the size can be tested without a real terminal. It
// returns the width (columns) and height (rows).
var getSizeFn = func() (width, height int) {
	ws, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 || ws.Row == 0 {
		return defaultWidth, defaultHeight
	}
	return int(ws.Col), int(ws.Row)
}

// Run executes ttysize.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[w] [h]", stdio.Err).WithHelp(command.Help{
		Description: "Print the terminal width and height in character cells, separated by a space. " +
			"Given the argument 'w' or 'h', print only the width or the height, in the order the " +
			"arguments appear. When the size cannot be determined, 80 and 24 are used.",
		Examples: []command.Example{
			{Command: "ttysize", Explain: "Print 'WIDTH HEIGHT'."},
			{Command: "ttysize w", Explain: "Print just the width."},
		},
		ExitStatus: "0  always.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	width, height := getSizeFn()

	rest := fs.Args()
	if len(rest) == 0 {
		_, _ = fmt.Fprintf(stdio.Out, "%d %d\n", width, height)
		return nil
	}

	var parts []string
	for _, a := range rest {
		switch a {
		case "w":
			parts = append(parts, fmt.Sprint(width))
		case "h":
			parts = append(parts, fmt.Sprint(height))
		default:
			return command.Failuref("unknown argument: %q (use w or h)", a)
		}
	}
	_, _ = fmt.Fprintln(stdio.Out, strings.Join(parts, " "))
	return nil
}
