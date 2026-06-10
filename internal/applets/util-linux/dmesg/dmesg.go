// Package dmesg implements the dmesg applet: print the kernel ring buffer.
package dmesg

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the dmesg applet.
type Command struct{}

// New returns a dmesg command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dmesg" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print or control the kernel ring buffer" }

// syslog actions (klogctl).
const (
	syslogActionReadAll = 3
)

// readKernelLog is indirected so the parser can be tested with a fixture.
var readKernelLog = func() ([]byte, error) {
	buf := make([]byte, 1<<20)
	n, err := unix.Klogctl(syslogActionReadAll, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Run executes dmesg.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-r]", stdio.Err).WithHelp(command.Help{
		Description: "Print the messages in the kernel ring buffer. Each line's syslog priority prefix " +
			"(<N>) is stripped unless -r (raw) is given. Reading the buffer may require privilege.",
		Examples: []command.Example{
			{Command: "dmesg", Explain: "Show the kernel messages."},
			{Command: "dmesg -r", Explain: "Show them with the raw priority prefixes."},
		},
		ExitStatus: "0  success.\n1  the kernel buffer could not be read.",
	})
	raw := fs.BoolP("raw", "r", false, "print the raw buffer, keeping priority prefixes")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	data, err := readKernelLog()
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "dmesg: read kernel buffer failed: %v\n", err)
		return command.SilentFailure()
	}

	_, _ = fmt.Fprint(stdio.Out, format(string(data), *raw))
	return nil
}

// format strips the syslog priority prefix from each non-empty line unless raw.
func format(s string, raw bool) string {
	if raw {
		return s
	}
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			continue
		}
		b.WriteString(stripPriority(line))
		b.WriteByte('\n')
	}
	return b.String()
}

// stripPriority removes a leading "<N>" syslog priority marker.
func stripPriority(line string) string {
	if strings.HasPrefix(line, "<") {
		if i := strings.IndexByte(line, '>'); i >= 0 {
			return line[i+1:]
		}
	}
	return line
}
