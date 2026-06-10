// Package rdate implements the rdate applet: get (and optionally set) the time
// from a remote host using the RFC 868 Time Protocol.
package rdate

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the rdate applet.
type Command struct{}

// New returns an rdate command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rdate" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Get the time from a remote host (RFC 868)" }

// epochOffset is the seconds between the RFC 868 epoch (1900-01-01) and the Unix
// epoch (1970-01-01).
const epochOffset = 2208988800

// Injected so the network and clock are controllable in tests.
var (
	port    = "37"
	timeout = 10 * time.Second
	setTime = func(t time.Time) error {
		tv := unixTimeval(t)
		return setSystemTime(&tv)
	}
)

// Run executes rdate.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-p] [-s] HOST", stdio.Err).WithHelp(command.Help{
		Description: "Connect to HOST's RFC 868 time service (TCP port 37) and read the current time. " +
			"With -p (the default) print it; with -s also set the system clock (which requires " +
			"privilege).",
		Examples: []command.Example{
			{Command: "rdate time.example.com", Explain: "Print the remote time."},
			{Command: "rdate -s time.example.com", Explain: "Set the local clock from the remote time."},
		},
		ExitStatus: "0  success.\n1  the host could not be reached or the time could not be set.",
	})
	printIt := fs.BoolP("print", "p", false, "print the remote time (the default)")
	setIt := fs.BoolP("set", "s", false, "set the system clock from the remote time")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	hosts := fs.Args()
	if len(hosts) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "rdate: a host is required")
		return command.SilentFailure()
	}

	remote, err := fetch(hosts[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "rdate: %v\n", err)
		return command.SilentFailure()
	}

	if *setIt {
		if err := setTime(remote); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "rdate: cannot set the system clock: %v\n", err)
			return command.SilentFailure()
		}
	}
	if *printIt || !*setIt {
		_, _ = fmt.Fprintln(stdio.Out, remote.Format("Mon Jan _2 15:04:05 2006"))
	}
	return nil
}

// fetch reads the time from host's RFC 868 service.
func fetch(host string) (time.Time, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return time.Time{}, err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	var buf [4]byte
	if _, err := io.ReadFull(conn, buf[:]); err != nil {
		return time.Time{}, err
	}
	secs := binary.BigEndian.Uint32(buf[:])
	return time.Unix(int64(secs)-epochOffset, 0), nil
}

// unixTimeval converts a time to a unix.Timeval for settimeofday(2).
func unixTimeval(t time.Time) unix.Timeval {
	return unix.Timeval{Sec: t.Unix(), Usec: int64(t.Nanosecond() / 1000)}
}

// setSystemTime sets the system clock (requires privilege).
func setSystemTime(tv *unix.Timeval) error { return unix.Settimeofday(tv) }
