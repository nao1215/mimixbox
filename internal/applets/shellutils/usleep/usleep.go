// Package usleep implements the usleep applet: pause for a number of
// microseconds.
package usleep

import (
	"context"
	"strconv"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the usleep applet.
type Command struct{}

// New returns a usleep command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "usleep" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Pause for N microseconds" }

// sleep is time.Sleep, indirected so tests do not actually wait.
var sleep = time.Sleep

// Run executes usleep.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[N]", stdio.Err).WithHelp(command.Help{
		Description: "Pause for N microseconds (default 0). N must be a non-negative integer.",
		Examples: []command.Example{
			{Command: "usleep 500000", Explain: "Pause for half a second."},
		},
		ExitStatus: "0  success.\n1  N was not a valid non-negative integer.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	micro := int64(0)
	if rest := fs.Args(); len(rest) > 0 {
		micro, err = strconv.ParseInt(rest[0], 10, 64)
		if err != nil || micro < 0 {
			return command.Failuref("invalid number of microseconds: %q", rest[0])
		}
	}
	sleep(time.Duration(micro) * time.Microsecond)
	return nil
}
