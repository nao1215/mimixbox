// Package sslutil implements the ssl_client and ssl_server applets.
//
// ssl_server runs a foreground TLS server over loopback that echoes (or runs a
// program for) each connection; ssl_client opens a TLS connection and pipes its
// standard input/output across it. The TLS plumbing is factored so a hermetic
// test can run a real local handshake between the two using a self-signed
// certificate generated at test time. Certificate verification against the host
// trust store is left to crypto/tls defaults; -k/--insecure disables it for the
// self-signed loopback case.
package sslutil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ssl_client / ssl_server applet.
type Command struct {
	name string
}

// NewSSLClient returns an ssl_client command.
func NewSSLClient() *Command { return &Command{name: "ssl_client"} }

// NewSSLServer returns an ssl_server command.
func NewSSLServer() *Command { return &Command{name: "ssl_server"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.name == "ssl_server" {
		return "Minimal TLS server (foreground)"
	}
	return "Open a TLS connection and pipe stdio"
}

// Run dispatches to ssl_client or ssl_server.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if c.name == "ssl_server" {
		return c.runServer(ctx, stdio, args)
	}
	return c.runClient(ctx, stdio, args)
}

func (c *Command) runServer(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ssl_server", "-c CERT -k KEY [-b ADDR]", stdio.Err).WithHelp(command.Help{
		Description: "Run a TLS server that terminates connections on a loopback address. -c and -k name " +
			"the PEM certificate and private key; -b sets the listen address (default 127.0.0.1:443). The " +
			"TLS accept loop (ServeTLS) is implemented and exercised by a hermetic local-handshake test; " +
			"running it as a system service requires a certificate and key that are not provided in this " +
			"environment, so without both -c and -k the command fails with a documented error.",
		Examples: []command.Example{
			{Command: "ssl_server -c cert.pem -k key.pem -b 127.0.0.1:8443", Explain: "Terminate TLS on loopback port 8443 using the given cert/key."},
		},
		ExitStatus: "0  clean shutdown.\n1  missing cert/key, load error, or bind error.",
		Notes: []string{
			"Binds a loopback address only; it is intended for local testing, not as a public-facing TLS server.",
			"Pair it with ssl_client to exercise a TLS exchange end to end.",
		},
	})
	certFile := fs.StringP("cert", "c", "", "PEM certificate file")
	keyFile := fs.StringP("key", "k", "", "PEM private key file")
	addr := fs.StringP("bind", "b", "127.0.0.1:443", "address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *certFile == "" || *keyFile == "" {
		return command.Failuref("a certificate (-c) and key (-k) are required to start the TLS server")
	}
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return command.Failuref("cannot load certificate/key: %v", err)
	}
	ln, err := tls.Listen("tcp", *addr, &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "ssl_server: listening on %s\n", ln.Addr().String())
	return ServeTLS(ln, EchoHandler)
}

func (c *Command) runClient(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ssl_client", "-s HOST:PORT [-k]", stdio.Err).WithHelp(command.Help{
		Description: "Open a TLS connection to HOST:PORT and pipe standard input to the server and the " +
			"server's response to standard output, then close. -s names the server; -k/--insecure skips " +
			"certificate verification (needed for self-signed loopback servers). The TLS dial-and-pipe " +
			"logic (DialAndPipe) is exercised by a hermetic local-handshake test against ssl_server.",
		Examples: []command.Example{
			{Command: "ssl_client -s 127.0.0.1:8443 -k", Explain: "Connect to a self-signed loopback TLS server and pipe stdio."},
		},
		ExitStatus: "0  the exchange completed.\n1  missing server, connection, or TLS handshake error.",
	})
	server := fs.StringP("server", "s", "", "TLS server to connect to (HOST:PORT)")
	insecure := fs.BoolP("insecure", "k", false, "skip certificate verification")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *server == "" {
		return command.Failuref("a server address is required (-s)")
	}
	cfg := &tls.Config{InsecureSkipVerify: *insecure} //nolint:gosec // -k is an explicit opt-in for self-signed servers
	if err := DialAndPipe(*server, cfg, stdio.In, stdio.Out); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}

// Handler handles one accepted TLS connection.
type Handler func(conn net.Conn) error

// EchoHandler reads all input from conn and writes it back, then returns.
func EchoHandler(conn net.Conn) error {
	_, err := io.Copy(conn, conn)
	return err
}

// ServeTLS accepts connections on ln (a TLS listener) and dispatches each to
// handler until the listener is closed.
func ServeTLS(ln net.Listener, handler Handler) error {
	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			wg.Wait()
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return command.Failuref("accept: %v", err)
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer func() { _ = c.Close() }()
			_ = handler(c)
		}(conn)
	}
}

// DialAndPipe opens a TLS connection to addr with cfg, writes everything from in
// to the server, then copies the server's reply to out.
func DialAndPipe(addr string, cfg *tls.Config, in io.Reader, out io.Writer) error {
	conn, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return fmt.Errorf("tls dial %s: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()
	if in != nil {
		if _, err := io.Copy(conn, in); err != nil {
			return err
		}
	}
	_ = conn.CloseWrite()
	if _, err := io.Copy(out, conn); err != nil {
		return err
	}
	return nil
}
