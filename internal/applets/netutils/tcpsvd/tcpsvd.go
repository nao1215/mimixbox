// Package tcpsvd implements the tcpsvd and udpsvd applets: accept connections on
// a TCP or UDP socket and run a program for each, wiring the socket to the
// program's stdin/stdout (the BusyBox tcpsvd/udpsvd model).
//
// The accept loop and per-connection child management common to both protocols
// live in the shared supervisor core (supervisor.go); this file holds the
// command entrypoints and the protocol-specific socket setup (TCP versus UDP),
// keeping the per-connection program execution separate so the loop can be
// tested over loopback with an injected handler instead of forking a real
// process.
package tcpsvd

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the tcpsvd / udpsvd applet.
type Command struct {
	name    string
	network string // "tcp" or "udp"
}

// NewTcpsvd returns a tcpsvd command.
func NewTcpsvd() *Command { return &Command{name: "tcpsvd", network: "tcp"} }

// NewUdpsvd returns a udpsvd command.
func NewUdpsvd() *Command { return &Command{name: "udpsvd", network: "udp"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.network == "udp" {
		return "Accept UDP datagrams and run a program for each"
	}
	return "Accept TCP connections and run a program for each"
}

// Run executes tcpsvd/udpsvd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-v] IP PORT PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Bind to IP:PORT and, for each incoming connection (tcpsvd) or datagram (udpsvd), run " +
			"PROG with the socket connected to its standard input and output. IP may be 0 to bind every " +
			"interface; for hermetic use bind a loopback address such as 127.0.0.1. The server runs in the " +
			"foreground until its context is cancelled, then stops accepting and returns cleanly.",
		Examples: []command.Example{
			{Command: "tcpsvd -v 127.0.0.1 8001 cat", Explain: "Echo each TCP connection's input back to it via cat."},
			{Command: "udpsvd 127.0.0.1 9001 cat", Explain: "Run cat for each UDP datagram on loopback port 9001."},
		},
		ExitStatus: "0  clean shutdown.\n1  bad arguments, bind error, or the program could not be started.",
		Notes: []string{
			"Connection concurrency limits (-c), user switching (-u), and DNS logging are not implemented in this slice.",
		},
	})
	verbose := fs.BoolP("verbose", "v", false, "log each connection to stderr")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 3 {
		return command.Failuref("usage: %s IP PORT PROG [ARG...]", c.name)
	}
	ip, portStr, prog := rest[0], rest[1], rest[2]
	progArgs := rest[3:]

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return command.Failuref("invalid port %q", portStr)
	}
	if ip == "0" {
		ip = "0.0.0.0"
	}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	handler := execHandler(ctx, prog, progArgs)
	if c.network == "udp" {
		return c.serveUDP(ctx, stdio, addr, *verbose, handler)
	}
	return c.serveTCP(ctx, stdio, addr, *verbose, handler)
}

// execHandler returns a ConnHandler that runs prog with the connection wired to
// the program's stdin/stdout.
func execHandler(ctx context.Context, prog string, args []string) ConnHandler {
	return func(conn net.Conn) error {
		cmd := exec.CommandContext(ctx, prog, args...)
		cmd.Stdin = conn
		cmd.Stdout = conn
		return cmd.Run()
	}
}

// serveTCP binds a TCP listener and dispatches each accepted connection to
// handler. It is exported behavior through Run but kept here so ServeTCP can be
// reused by tests via the package-level helper.
func (c *Command) serveTCP(ctx context.Context, stdio command.IO, addr string, verbose bool, handler ConnHandler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", addr, err)
	}
	return ServeTCP(ctx, ln, stdio, verbose, handler)
}

// ServeTCP runs the TCP accept loop on ln until ctx is cancelled.
func ServeTCP(ctx context.Context, ln net.Listener, stdio command.IO, verbose bool, handler ConnHandler) error {
	return supervisor{
		sock:     ln,
		verbose:  verbose,
		logLine:  func(conn net.Conn) string { return fmt.Sprintf("tcpsvd: connection from %s", conn.RemoteAddr()) },
		errLabel: "accept",
		accept: func() (net.Conn, bool, error) {
			conn, err := ln.Accept()
			return conn, true, err
		},
		handler: handler,
	}.serve(ctx, stdio)
}

// serveUDP binds a UDP socket and, for each datagram, invokes handler with a
// connection backed by that single datagram exchange.
func (c *Command) serveUDP(ctx context.Context, stdio command.IO, addr string, verbose bool, handler ConnHandler) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return command.Failuref("resolve %s: %v", addr, err)
	}
	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", addr, err)
	}
	return ServeUDP(ctx, pc, stdio, verbose, handler)
}

// ServeUDP runs the UDP receive loop on pc until ctx is cancelled. Each datagram
// is delivered to handler through a udpConn that reads the datagram payload and
// writes replies back to the sender.
func ServeUDP(ctx context.Context, pc *net.UDPConn, stdio command.IO, verbose bool, handler ConnHandler) error {
	buf := make([]byte, 64*1024)
	return supervisor{
		sock:     pc,
		verbose:  verbose,
		logLine:  func(conn net.Conn) string { return fmt.Sprintf("udpsvd: datagram from %s", conn.RemoteAddr()) },
		errLabel: "read",
		accept: func() (net.Conn, bool, error) {
			n, raddr, err := pc.ReadFromUDP(buf)
			if err != nil {
				return nil, false, err
			}
			payload := make([]byte, n)
			copy(payload, buf[:n])
			return newUDPConn(pc, raddr, payload), false, nil
		},
		handler: handler,
	}.serve(ctx, stdio)
}
