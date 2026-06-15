// Package telnetd implements the telnetd applet: a minimal foreground telnet
// server that runs a program for each connection.
//
// Telnet IAC command stripping is a pure function so it can be table-tested, and
// the accept loop runs over loopback with an injectable session handler for
// hermetic integration tests instead of allocating a real PTY.
package telnetd

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Telnet protocol bytes.
const (
	iac = 255 // Interpret As Command
	sb  = 250 // subnegotiation begin
	se  = 240 // subnegotiation end
)

// Command is the telnetd applet.
type Command struct{}

// New returns a telnetd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "telnetd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Minimal telnet server (foreground)" }

// Run executes telnetd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-b ADDR] [-l PROG [ARG...]]", stdio.Err).WithHelp(command.Help{
		Description: "Run a telnet server that executes a program for each connection, wiring the (IAC-" +
			"filtered) network stream to the program's standard input and output. -f keeps telnetd in the " +
			"foreground; -b sets the listen address (default 127.0.0.1:23); -l selects the login program " +
			"and its arguments (default: /bin/sh). The server runs until its context is cancelled.",
		Examples: []command.Example{
			{Command: "telnetd -f -b 127.0.0.1:2323 -l /bin/cat", Explain: "Echo each telnet session via cat on loopback port 2323."},
		},
		ExitStatus: "0  clean shutdown.\n1  bad arguments or bind error.",
		Notes: []string{
			"Pseudo-terminal allocation and option negotiation beyond IAC stripping are not implemented in this slice.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("bind", "b", "127.0.0.1:23", "address to listen on (HOST:PORT)")
	login := fs.StringP("login", "l", "/bin/sh", "program to run for each session")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "telnetd: listening on %s, login=%s\n", ln.Addr().String(), *login)
	handler := execSession(ctx, *login, fs.Args())
	return Serve(ctx, ln, handler)
}

// Session handles a single accepted telnet connection.
type Session func(conn net.Conn) error

// execSession returns a Session that runs prog with the IAC-filtered network
// stream wired to the program's stdin/stdout.
func execSession(ctx context.Context, prog string, args []string) Session {
	return func(conn net.Conn) error {
		cmd := exec.CommandContext(ctx, prog, args...)
		cmd.Stdin = NewReader(conn)
		cmd.Stdout = conn
		return cmd.Run()
	}
}

// Serve runs the telnet accept loop on ln until ctx is cancelled.
func Serve(ctx context.Context, ln net.Listener, handler Session) error {
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
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			defer func() { _ = c.Close() }()
			_ = handler(c)
		}(conn)
	}
}

// StripIAC removes telnet IAC command sequences from p, returning only the data
// bytes. A 3-byte command (IAC WILL/WONT/DO/DONT OPTION) and SB...SE
// subnegotiations are dropped, a bare IAC IAC pair collapses to one 0xFF data
// byte, and any incomplete trailing sequence is discarded.
func StripIAC(p []byte) []byte {
	data, _ := stripIACKeepPartial(p)
	return data
}
