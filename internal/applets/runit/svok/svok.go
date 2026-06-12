// Package svok implements the svok applet: check whether a service directory is
// under active supervision.
package svok

import (
	"context"
	"os"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the svok applet.
type Command struct{}

// New returns a svok command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "svok" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Check if a service is supervised" }

// Run executes svok.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DIR", stdio.Err).WithHelp(command.Help{
		Description: "Check whether the service directory DIR is under active supervision, by testing " +
			"for its supervise/ok control file (which runsv keeps open while supervising). Exit 0 if " +
			"the service is supervised, or 100 if it is not.",
		Examples: []command.Example{
			{Command: "svok /etc/service/nginx", Explain: "Succeed if nginx is supervised."},
		},
		ExitStatus: "0   the service is supervised.\n100 the service is not supervised.\n1   no directory was given.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a service directory is required")
	}

	if _, err := os.Stat(filepath.Join(rest[0], "supervise", "ok")); err != nil {
		return &command.ExitError{Code: 100}
	}
	return nil
}
