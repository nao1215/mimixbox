// Package logread implements the logread applet: print the contents of the
// system log.
package logread

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the logread applet.
type Command struct{}

// New returns a logread command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "logread" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the system log" }

// logCandidates are tried in order; tests point this at a fixture.
var logCandidates = []string{"/var/log/messages", "/var/log/syslog"}

// Run executes logread.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[FILE]", stdio.Err).WithHelp(command.Help{
		Description: "Print the system log. With no argument the first readable of /var/log/messages " +
			"and /var/log/syslog is shown; a FILE argument overrides that. Following the log (-f) is " +
			"not supported.",
		Examples: []command.Example{
			{Command: "logread", Explain: "Print the system log."},
		},
		ExitStatus: "0  the log was printed.\n1  no readable log was found.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	candidates := logCandidates
	if rest := fs.Args(); len(rest) > 0 {
		candidates = rest
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path) //nolint:gosec // user-named or well-known log path
		if err != nil {
			continue
		}
		_, _ = stdio.Out.Write(data)
		return nil
	}

	_, _ = fmt.Fprintln(stdio.Err, "logread: no readable system log was found")
	return command.SilentFailure()
}
