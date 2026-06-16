// Package sslutil implements the ssl_client and ssl_server applets.
//
// ssl_server runs a foreground TLS server over loopback that echoes (or runs a
// program for) each connection; ssl_client opens a TLS connection and pipes its
// standard input/output across it. The TLS plumbing is factored so a hermetic
// test can run a real local handshake between the two using a self-signed
// certificate generated at test time. Certificate verification against the host
// trust store is left to crypto/tls defaults; -k/--insecure disables it for the
// self-signed loopback case.
//
// tls.go holds the shared TLS/session setup (cert/key loading, tls.Config
// construction, and the accept/dial connection plumbing); ssl_server.go and
// ssl_client.go hold the per-command entrypoints.
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

// ServerConfig loads the PEM certificate and key and returns a tls.Config that
// presents them. It is the shared server-side TLS setup used by ssl_server.
func ServerConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("cannot load certificate/key: %w", err)
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

// ClientConfig returns the client-side tls.Config. When insecure is true it
// skips certificate verification, which the -k/--insecure flag opts into for
// self-signed loopback servers.
func ClientConfig(insecure bool) *tls.Config {
	return &tls.Config{InsecureSkipVerify: insecure} //nolint:gosec // -k is an explicit opt-in for self-signed servers
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
