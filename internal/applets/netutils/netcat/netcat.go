// Package netcat implements the netcat applet: a thin compatibility front over
// the existing nc applet. BusyBox ships both names for the same tool, so netcat
// reuses the nc implementation verbatim and only differs in the command name
// reported in usage and the applet list.
package netcat

import (
	"context"

	"github.com/nao1215/mimixbox/internal/applets/netutils/nc"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the netcat applet.
type Command struct {
	delegate *nc.Command
}

// New returns a netcat command that delegates to nc.
func New() *Command { return &Command{delegate: nc.New()} }

// Name returns the command name.
func (c *Command) Name() string { return "netcat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Read and write data across network connections (alias of nc)"
}

// Run executes netcat by forwarding to the nc implementation.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	return c.delegate.Run(ctx, stdio, args)
}
