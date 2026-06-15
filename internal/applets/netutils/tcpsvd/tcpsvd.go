// Package tcpsvd implements the tcpsvd and udpsvd applets: accept connections on
// a TCP or UDP socket and run a program for each, wiring the socket to the
// program's stdin/stdout (the BusyBox tcpsvd/udpsvd model).
//
// The accept loop and the per-connection program execution are separated so the
// loop can be tested over loopback with an injected handler instead of forking a
// real process.
package tcpsvd

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

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

// ConnHandler handles one accepted connection (TCP) or datagram exchange (UDP).
type ConnHandler func(conn net.Conn) error

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

// serveTCP accepts TCP connections until ctx is cancelled, dispatching each to
// handler. It is exported behavior through Run but kept here so ListenAndServe
// can be reused by tests via the package-level helper.
func (c *Command) serveTCP(ctx context.Context, stdio command.IO, addr string, verbose bool, handler ConnHandler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", addr, err)
	}
	return ServeTCP(ctx, ln, stdio, verbose, handler)
}

// ServeTCP runs the TCP accept loop on ln until ctx is cancelled.
func ServeTCP(ctx context.Context, ln net.Listener, stdio command.IO, verbose bool, handler ConnHandler) error {
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				wg.Wait()
				return nil
			}
			return command.Failuref("accept: %v", err)
		}
		if verbose {
			_, _ = fmt.Fprintf(stdio.Err, "tcpsvd: connection from %s\n", conn.RemoteAddr())
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer func() { _ = c.Close() }()
			_ = handler(c)
		}(conn)
	}
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
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()

	buf := make([]byte, 64*1024)
	for {
		n, raddr, err := pc.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return command.Failuref("read: %v", err)
		}
		if verbose {
			_, _ = fmt.Fprintf(stdio.Err, "udpsvd: datagram from %s\n", raddr)
		}
		payload := make([]byte, n)
		copy(payload, buf[:n])
		conn := newUDPConn(pc, raddr, payload)
		_ = handler(conn)
	}
}

// udpConn adapts a single received datagram to the net.Conn interface so the
// same ConnHandler works for both TCP and UDP. Read yields the datagram payload
// once; Write sends a reply datagram back to the original sender.
type udpConn struct {
	pc      *net.UDPConn
	raddr   *net.UDPAddr
	payload []byte
	off     int
}

func newUDPConn(pc *net.UDPConn, raddr *net.UDPAddr, payload []byte) *udpConn {
	return &udpConn{pc: pc, raddr: raddr, payload: payload}
}

func (u *udpConn) Read(p []byte) (int, error) {
	if u.off >= len(u.payload) {
		return 0, io.EOF
	}
	n := copy(p, u.payload[u.off:])
	u.off += n
	return n, nil
}

func (u *udpConn) Write(p []byte) (int, error)       { return u.pc.WriteToUDP(p, u.raddr) }
func (u *udpConn) Close() error                      { return nil }
func (u *udpConn) LocalAddr() net.Addr               { return u.pc.LocalAddr() }
func (u *udpConn) RemoteAddr() net.Addr              { return u.raddr }
func (u *udpConn) SetDeadline(_ time.Time) error     { return nil }
func (u *udpConn) SetReadDeadline(_ time.Time) error { return nil }
func (u *udpConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

var _ net.Conn = (*udpConn)(nil)
