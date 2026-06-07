// Package sync implements the sync applet: flush filesystem buffers so cached
// writes reach persistent storage.
package sync

import (
	"context"
	"fmt"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sync applet.
type Command struct{}

// New returns a sync command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sync" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Synchronize cached writes to persistent storage" }

// Run executes sync. In its basic form sync takes no operands; it flushes the
// filesystem buffers via syscall.Sync().
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if err := sync(); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "sync: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// sync flushes filesystem buffers to persistent storage.
func sync() (err error) {
	// syscall.Sync returns no value; recover defends against the rare platform
	// where the call could panic, keeping Run's error path meaningful.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	syscall.Sync()
	return nil
}
