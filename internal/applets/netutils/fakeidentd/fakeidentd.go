// Package fakeidentd implements the fakeidentd applet: a minimal RFC 1413 ident
// daemon that answers every query with a fixed user name.
//
// The protocol handling (parsing "PORT,PORT" and formatting the USERID reply) is
// a pure function so it can be unit-tested, and the foreground server runs over
// loopback for hermetic integration tests.
package fakeidentd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the fakeidentd applet.
type Command struct{}

// New returns a fakeidentd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "fakeidentd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Answer ident (RFC 1413) queries with a fixed user" }

// Run executes fakeidentd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-b ADDR] [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Run a fake RFC 1413 ident daemon that replies to every well-formed query with the " +
			"same user name (default: nobody). -f keeps the daemon in the foreground; -b sets the listen " +
			"address (default 127.0.0.1:113). The optional USER operand overrides the reported name. The " +
			"daemon runs until its context is cancelled, then stops accepting and returns cleanly.",
		Examples: []command.Example{
			{Command: "fakeidentd -f -b 127.0.0.1:1130 alice", Explain: "Answer ident queries on loopback port 1130 as 'alice'."},
		},
		ExitStatus: "0  clean shutdown.\n1  bind error or unsupported mode.",
		Notes: []string{
			"Foreground mode (-f) is implemented; inetd-style/background operation is not implemented in this slice.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("bind", "b", "127.0.0.1:113", "address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}

	user := "nobody"
	if rest := fs.Args(); len(rest) > 0 {
		user = rest[0]
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "fakeidentd: listening on %s as %q\n", ln.Addr().String(), user)
	return Serve(ctx, ln, user)
}

// Serve runs the ident accept loop on ln until ctx is cancelled.
func Serve(ctx context.Context, ln net.Listener, user string) error {
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
			handle(c, user)
		}(conn)
	}
}

// handle reads one ident query line and writes the reply.
func handle(conn net.Conn, user string) {
	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return
	}
	_, _ = conn.Write([]byte(Reply(line, user)))
}

// Reply builds the RFC 1413 response for query. A well-formed query is
// "SERVERPORT,CLIENTPORT"; the response echoes those ports and reports user as
// the USERID. A malformed query yields an INVALID-PORT error response.
func Reply(query, user string) string {
	q := strings.TrimSpace(query)
	parts := strings.SplitN(q, ",", 2)
	if len(parts) != 2 {
		return q + " : ERROR : INVALID-PORT\r\n"
	}
	sp := strings.TrimSpace(parts[0])
	cp := strings.TrimSpace(parts[1])
	if !validPort(sp) || !validPort(cp) {
		return q + " : ERROR : INVALID-PORT\r\n"
	}
	return fmt.Sprintf("%s , %s : USERID : UNIX : %s\r\n", sp, cp, user)
}

func validPort(s string) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1 && n <= 65535
}
