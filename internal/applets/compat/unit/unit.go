// Package unit implements a compatibility "unit" applet. In BusyBox, unit is the
// developer-facing libbb unit-test runner. MimixBox does not ship that suite, so
// rather than silently doing nothing this applet explains where MimixBox's tests
// live and exits non-zero.
package unit

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the unit applet.
type Command struct{}

// New returns a unit command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unit" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "BusyBox unit-test runner (not shipped by MimixBox)" }

// Run reports that MimixBox has no libbb unit-test suite and exits non-zero.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "BusyBox's unit applet runs its internal libbb unit tests. MimixBox does not " +
			"embed that suite, so this applet does nothing except point at MimixBox's own tests.",
		Notes: []string{
			"Run 'go test ./...' for unit tests and 'make test-e2e' for the ShellSpec suite.",
		},
		ExitStatus: "2  always, because there is no embedded unit-test suite to run.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	_, _ = fmt.Fprintln(stdio.Err, "unit: MimixBox ships no embedded unit-test suite; run 'go test ./...' and 'make test-e2e' instead")
	return &command.ExitError{Code: 2}
}
