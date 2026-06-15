// Package telnet implements the telnet applet: open a raw TCP connection to a
// host and port and shuttle data between the terminal and the connection. The
// dialer is injectable so tests drive it against a loopback fixture server. This
// is a line-oriented raw client; full TELNET option negotiation (IAC commands)
// is intentionally not implemented.
package telnet

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the telnet applet.
type Command struct{}

// New returns a telnet command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "telnet" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Connect to a host over TCP (raw, line-oriented)" }

// dial opens a connection to address; tests replace it with a loopback dialer.
var dial = func(address string) (net.Conn, error) { return net.Dial("tcp", address) }

// Run executes telnet.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "HOST [PORT]", stdio.Err).WithHelp(command.Help{
		Description: "Open a raw TCP connection to HOST on PORT (default 23) and copy standard input " +
			"to the connection and the connection's data to standard output until either side closes. " +
			"This is a raw, line-oriented client: TELNET option negotiation (IAC sequences) is not " +
			"performed, so it behaves like a minimal nc for interactive line protocols.",
		Examples: []command.Example{
			{Command: "telnet localhost 25", Explain: "Speak to an SMTP server on port 25."},
			{Command: "telnet example.test", Explain: "Connect on the default telnet port 23."},
		},
		ExitStatus: "0  the session ended normally.\n" +
			"1  bad arguments or the connection failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	host, port, err := hostPort(fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}

	conn, err := dial(net.JoinHostPort(host, port))
	if err != nil {
		return command.Failuref("cannot connect to %s:%s: %v", host, port, err)
	}
	defer func() { _ = conn.Close() }()
	return shuttle(conn, stdio.In, stdio.Out)
}

// hostPort extracts HOST and PORT (default 23) from the operands.
func hostPort(operands []string) (host, port string, err error) {
	switch len(operands) {
	case 1:
		return operands[0], "23", nil
	case 2:
		return operands[0], operands[1], nil
	default:
		return "", "", fmt.Errorf("usage: telnet HOST [PORT]")
	}
}

// halfCloser is implemented by *net.TCPConn; closing the write half lets the
// peer observe EOF once stdin is exhausted.
type halfCloser interface{ CloseWrite() error }

// shuttle copies in->conn and conn->out concurrently and returns when the
// connection-to-output direction completes.
func shuttle(conn net.Conn, in io.Reader, out io.Writer) error {
	go func() {
		_, _ = io.Copy(conn, in)
		if cw, ok := conn.(halfCloser); ok {
			_ = cw.CloseWrite()
		}
	}()
	if _, err := io.Copy(out, conn); err != nil {
		return command.Failure(err)
	}
	return nil
}
