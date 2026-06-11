// Package uevent implements the uevent applet: monitor kernel uevents from the
// netlink socket and print each one until interrupted.
package uevent

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the uevent applet.
type Command struct{}

// New returns a uevent command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "uevent" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Monitor kernel uevents" }

// ueventSource yields kernel uevent messages.
type ueventSource interface {
	Recv() ([]byte, error)
	Close() error
}

// dialFn is indirected so the monitor can be tested without a netlink socket.
var dialFn = func() (ueventSource, error) { return openNetlink() }

// Run executes uevent.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Listen on the kernel uevent netlink socket and print each event as 'ACTION " +
			"DEVPATH' as devices are added, removed, or changed, until interrupted. Opening the " +
			"netlink socket requires privilege.",
		Examples: []command.Example{
			{Command: "uevent", Explain: "Print kernel uevents as they arrive."},
		},
		ExitStatus: "0  the monitor stopped cleanly.\n1  the netlink socket could not be opened.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	src, err := dialFn()
	if err != nil {
		return command.Failuref("cannot open the uevent netlink socket: %v", err)
	}
	defer func() { _ = src.Close() }()

	// Closing the source on cancellation unblocks Recv.
	go func() {
		<-ctx.Done()
		_ = src.Close()
	}()

	for {
		msg, err := src.Recv()
		if err != nil {
			return nil // closed via cancellation, or the source ended
		}
		action, devpath := parseEvent(msg)
		if action == "" {
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s %s\n", action, devpath)
	}
}

// parseEvent extracts the action and device path from a uevent message, whose
// header is "ACTION@DEVPATH" followed by NUL-separated KEY=VALUE pairs.
func parseEvent(msg []byte) (action, devpath string) {
	header := msg
	if i := bytes.IndexByte(msg, 0); i >= 0 {
		header = msg[:i]
	}
	a, d, found := strings.Cut(string(header), "@")
	if !found {
		return "", ""
	}
	return a, d
}

// netlinkSource reads uevents from a NETLINK_KOBJECT_UEVENT socket.
type netlinkSource struct{ fd int }

func openNetlink() (ueventSource, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_KOBJECT_UEVENT)
	if err != nil {
		return nil, err
	}
	addr := &unix.SockaddrNetlink{Family: unix.AF_NETLINK, Groups: 1}
	if err := unix.Bind(fd, addr); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return &netlinkSource{fd: fd}, nil
}

func (s *netlinkSource) Recv() ([]byte, error) {
	buf := make([]byte, 8192)
	n, _, err := unix.Recvfrom(s.fd, buf, 0)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *netlinkSource) Close() error { return unix.Close(s.fd) }
