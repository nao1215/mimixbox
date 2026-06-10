// Package klogd implements the klogd applet: read the kernel ring buffer and
// forward each message to the system log (one-shot mode).
package klogd

import (
	"context"
	"log/syslog"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the klogd applet.
type Command struct{}

// New returns a klogd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "klogd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Forward kernel messages to the system log" }

// syslogActionReadAll is SYSLOG_ACTION_READ_ALL for klogctl(2).
const syslogActionReadAll = 3

// readKernelLog is indirected so the forwarding can be tested without root.
var readKernelLog = func() ([]byte, error) {
	buf := make([]byte, 1<<20)
	n, err := unix.Klogctl(syslogActionReadAll, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// logFunc is indirected so it is testable without a running syslogd.
var logFunc = func(p syslog.Priority, msg string) error {
	w, err := syslog.New(p, "kernel")
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()
	_, err = w.Write([]byte(msg))
	return err
}

// Run executes klogd in one-shot mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-o]", stdio.Err).WithHelp(command.Help{
		Description: "Read the kernel ring buffer once and forward each message to the system log " +
			"with the LOG_KERN facility, preserving each message's priority. Only one-shot mode is " +
			"implemented; the continuously-following daemon is not.",
		Examples: []command.Example{
			{Command: "klogd -o", Explain: "Forward the current kernel messages and exit."},
		},
		ExitStatus: "0  the messages were forwarded.\n1  the kernel buffer or syslog was unreachable.",
	})
	_ = fs.BoolP("once", "o", false, "read the buffer once and exit (the only supported mode)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	data, err := readKernelLog()
	if err != nil {
		return command.Failuref("cannot read the kernel buffer: %v", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		prio, text := splitPriority(line)
		if err := logFunc(prio, text); err != nil {
			return command.Failuref("cannot forward to syslog: %v", err)
		}
	}
	return nil
}

// splitPriority parses a leading "<N>" kernel priority prefix, returning the
// LOG_KERN-facility priority and the message text. Without a prefix it defaults
// to LOG_KERN|LOG_INFO.
func splitPriority(line string) (syslog.Priority, string) {
	prio := syslog.LOG_KERN | syslog.LOG_INFO
	if strings.HasPrefix(line, "<") {
		if end := strings.IndexByte(line, '>'); end > 1 {
			if n, err := strconv.Atoi(line[1:end]); err == nil {
				prio = syslog.LOG_KERN | syslog.Priority(n&0x7)
				return prio, line[end+1:]
			}
		}
	}
	return prio, line
}
