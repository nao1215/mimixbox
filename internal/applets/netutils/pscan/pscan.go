// Package pscan implements the pscan applet: scan a range of TCP ports on a host
// and report which are open. The probe is an injectable TCP-connect function so
// tests can drive it against loopback listeners (or a fake) without external
// network access.
package pscan

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the pscan applet.
type Command struct{}

// New returns a pscan command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pscan" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Scan a range of TCP ports on a host" }

// probe reports whether a TCP connection to host:port succeeds within timeout.
// Tests replace it to avoid real connections.
var probe = func(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprint(port)), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Run executes pscan.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-p MIN] [-P MAX] [-t TIMEOUT_MS] HOST", stdio.Err).WithHelp(command.Help{
		Description: "Scan TCP ports MIN..MAX (inclusive) on HOST using non-blocking connect probes " +
			"and print each open port. MIN defaults to 1 and MAX to 1024. TIMEOUT_MS bounds each " +
			"probe (default 200ms). Only TCP connect scanning is performed; no raw sockets or root " +
			"privileges are required.",
		Examples: []command.Example{
			{Command: "pscan 127.0.0.1", Explain: "Scan ports 1-1024 on localhost."},
			{Command: "pscan -p 20 -P 25 127.0.0.1", Explain: "Scan a small range."},
		},
		ExitStatus: "0  the scan completed (open ports, if any, were printed).\n" +
			"1  bad arguments.",
	})
	min := fs.IntP("min-port", "p", 1, "lowest port to scan")
	max := fs.IntP("max-port", "P", 1024, "highest port to scan")
	timeoutMS := fs.IntP("timeout", "t", 200, "per-port connect timeout in milliseconds")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) != 1 {
		return command.Failuref("exactly one HOST is required")
	}
	host := operands[0]
	if *min < 1 || *max > 65535 || *min > *max {
		return command.Failuref("invalid port range %d-%d (must be within 1-65535 and min<=max)", *min, *max)
	}

	timeout := time.Duration(*timeoutMS) * time.Millisecond
	for port := *min; port <= *max; port++ {
		if probe(host, port, timeout) {
			fmt.Fprintf(stdio.Out, "%d: open\n", port)
		}
	}
	return nil
}
