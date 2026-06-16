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

// Run executes netcat by forwarding to the nc implementation. A leading --help
// renders netcat-named structured help instead of nc's so the usage and example
// lines match the invoked command.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if command.HandleHelpVersionWith(stdio, c.Name(), "[OPTION]... [HOST] [PORT]", command.Help{
		Description: "Read and write data across TCP or UDP network connections. netcat is an alias of " +
			"nc: it can open a connection to HOST PORT, or listen for one with -l.",
		Examples: []command.Example{
			{Command: "netcat -l 8080", Explain: "Listen for a TCP connection on port 8080."},
			{Command: "netcat example.com 80", Explain: "Connect to example.com on port 80."},
		},
		ExitStatus: "0  the connection completed successfully.\n1  a network or usage error occurred.",
	}, args) {
		return nil
	}
	return c.delegate.Run(ctx, stdio, args)
}
