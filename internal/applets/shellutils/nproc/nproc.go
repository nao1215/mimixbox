// Package nproc implements the nproc applet: print the number of processing
// units available to the current process.
package nproc

import (
	"context"
	"fmt"
	"runtime"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the nproc applet.
type Command struct{}

// New returns an nproc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nproc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the number of processing units available" }

// cpuCount reports the number of usable CPUs; tests replace it.
var cpuCount = runtime.NumCPU

// Run executes nproc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]...", stdio.Err)
	ignore := fs.Uint("ignore", 0, "if possible, exclude N processing units")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	n := cpuCount()
	n -= int(*ignore)
	if n < 1 {
		n = 1
	}
	if _, err := fmt.Fprintln(stdio.Out, n); err != nil {
		return command.Failure(err)
	}
	return nil
}
