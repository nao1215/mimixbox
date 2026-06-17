// Package nc implements the nc (netcat) applet: read and write data across TCP
// or UDP connections.
//
// This is a clean-room reimplementation. The original morrigan netcat was
// forked from vfedoroff/go-netcat (MIT License, Copyright (c) 2016 Vasily
// Fedoroff); the attribution is preserved here per the MIT terms even though no
// source is copied verbatim.
package nc

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/nao1215/mimixbox/internal/command"
)

// Network seams. These are package-level variables so tests can replace them
// with in-memory primitives (net.Pipe / fake packet conns) and exercise the
// connect/serve logic without binding real loopback sockets. Production wiring
// dials and listens for real.
var (
	dial         = func(network, address string) (net.Conn, error) { return net.Dial(network, address) }
	listen       = func(network, address string) (net.Listener, error) { return net.Listen(network, address) }
	listenPacket = func(network, address string) (net.PacketConn, error) { return net.ListenPacket(network, address) }
)

// Command is the nc applet.
type Command struct{}

// New returns an nc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read and write data across network connections" }

// Run executes nc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-u] [-l] [-p PORT] [HOST] [PORT]", stdio.Err).WithHelp(command.Help{
		Description: "Read and write data across a TCP (default) or UDP (-u) network connection. By default nc connects to HOST and PORT and shuttles data between the connection and the standard streams; with -l it listens for a single incoming connection instead.",
		Examples: []command.Example{
			{Command: "nc example.com 80", Explain: "Open a TCP connection to example.com on port 80."},
			{Command: "nc -l -p 8080", Explain: "Listen on TCP port 8080 for one incoming connection."},
		},
		ExitStatus: "0  the connection completed normally.\n1  the connection could not be established or failed.",
	})
	listen := fs.BoolP("listen", "l", false, "listen for an incoming connection")
	udp := fs.BoolP("udp", "u", false, "use UDP instead of TCP")
	port := fs.StringP("port", "p", "", "listen on (or source from) this port")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	network := "tcp"
	if *udp {
		network = "udp"
	}

	if *listen {
		addr, err := listenAddr(*port, fs.Args())
		if err != nil {
			return command.Failuref("%v", err)
		}
		return c.serve(stdio, network, addr)
	}

	host, p, err := dialAddr(*port, fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}
	return c.connect(stdio, network, net.JoinHostPort(host, p))
}

// listenAddr resolves the address to listen on from -p and the operands.
func listenAddr(portFlag string, operands []string) (string, error) {
	host, port := "", portFlag
	switch len(operands) {
	case 0:
	case 1:
		port = operands[0]
	default:
		host, port = operands[0], operands[1]
	}
	if port == "" {
		return "", fmt.Errorf("a port is required to listen")
	}
	return net.JoinHostPort(host, port), nil
}

// dialAddr resolves the host and port to connect to from the operands.
func dialAddr(portFlag string, operands []string) (string, string, error) {
	switch len(operands) {
	case 2:
		return operands[0], operands[1], nil
	case 1:
		if portFlag != "" {
			return operands[0], portFlag, nil
		}
		return "", "", fmt.Errorf("both HOST and PORT are required")
	default:
		return "", "", fmt.Errorf("both HOST and PORT are required")
	}
}

// connect dials address and shuttles data between the connection and the
// command's standard streams until either side closes.
func (c *Command) connect(stdio command.IO, network, address string) error {
	conn, err := dial(network, address)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer func() { _ = conn.Close() }()
	return shuttle(conn, stdio.In, stdio.Out)
}

// serve listens on address, accepts one connection and shuttles data over it.
func (c *Command) serve(stdio command.IO, network, address string) error {
	if network == "udp" {
		return c.serveUDP(stdio, address)
	}
	ln, err := listen(network, address)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer func() { _ = ln.Close() }()

	conn, err := ln.Accept()
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer func() { _ = conn.Close() }()
	return shuttle(conn, stdio.In, stdio.Out)
}

// serveUDP receives the first UDP payload and writes it to stdout.
func (c *Command) serveUDP(stdio command.IO, address string) error {
	pc, err := listenPacket("udp", address)
	if err != nil {
		return command.Failuref("%v", err)
	}
	defer func() { _ = pc.Close() }()

	buf := make([]byte, 64*1024)
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		return command.Failuref("%v", err)
	}
	if _, err := stdio.Out.Write(buf[:n]); err != nil {
		return command.Failure(err)
	}
	return nil
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
