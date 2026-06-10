// Package syslogd implements the syslogd applet: a minimal system-log daemon
// that receives messages on a Unix datagram socket and appends them to a file.
package syslogd

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the syslogd applet.
type Command struct{}

// New returns a syslogd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "syslogd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Minimal system logging daemon" }

// Run executes syslogd in the foreground until the context is cancelled.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-l SOCKET] [-O LOGFILE]", stdio.Err).WithHelp(command.Help{
		Description: "Receive log messages on a Unix datagram socket and append them to a log file, " +
			"running in the foreground until interrupted. -l sets the socket (default /dev/log) and " +
			"-O the log file (default /var/log/messages). The priority prefix is stripped from each " +
			"message. Background daemonisation and the shared-memory buffer are not implemented.",
		Examples: []command.Example{
			{Command: "syslogd -l /tmp/log -O /tmp/messages", Explain: "Log to a custom socket and file."},
		},
		ExitStatus: "0  the daemon stopped cleanly.\n1  the socket or log file could not be opened.",
	})
	socket := fs.StringP("socket", "l", "/dev/log", "datagram socket to listen on")
	logfile := fs.StringP("output", "O", "/var/log/messages", "file to append messages to")
	_ = fs.BoolP("foreground", "n", false, "run in the foreground (the only supported mode)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_ = os.Remove(*socket) // a stale socket would block ListenUnixgram
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: *socket, Net: "unixgram"})
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *socket, err)
	}
	defer func() { _ = conn.Close() }()
	defer func() { _ = os.Remove(*socket) }()

	out, err := os.OpenFile(*logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644) //nolint:gosec // user-named log
	if err != nil {
		return command.Failuref("cannot open %s: %v", *logfile, err)
	}
	defer func() { _ = out.Close() }()

	// Closing the connection when the context is cancelled unblocks Read.
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	buf := make([]byte, 16*1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil // connection closed via context cancellation
		}
		if _, err := fmt.Fprintln(out, stripPriority(string(buf[:n]))); err != nil {
			return command.Failuref("cannot write to %s: %v", *logfile, err)
		}
	}
}

// stripPriority removes a leading "<N>" syslog priority prefix (N numeric).
func stripPriority(msg string) string {
	if strings.HasPrefix(msg, "<") {
		if end := strings.IndexByte(msg, '>'); end > 1 && isDigits(msg[1:end]) {
			return strings.TrimRight(msg[end+1:], "\n")
		}
	}
	return strings.TrimRight(msg, "\n")
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
