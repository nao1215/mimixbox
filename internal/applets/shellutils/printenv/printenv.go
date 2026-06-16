// Package printenv implements the printenv applet: print all or named
// environment variables.
package printenv

import (
	"context"
	"fmt"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the printenv applet.
type Command struct{}

// New returns a printenv command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "printenv" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print environment variable" }

// Run executes printenv.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [VARIABLE]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the values of the named environment VARIABLEs, one per line. With no operand, " +
			"print every environment variable as NAME=VALUE.",
		Examples: []command.Example{
			{Command: "printenv", Explain: "Print all environment variables as NAME=VALUE."},
			{Command: "printenv HOME", Explain: "Print the value of the HOME variable."},
			{Command: "printenv -0 PATH", Explain: "Print PATH terminated by a NUL byte."},
		},
		ExitStatus: "0  all requested variables were set.\n1  one or more requested variables were unset.",
	})
	null := fs.BoolP("null", "0", false, "end each output line with NUL, not newline")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	end := byte('\n')
	if *null {
		end = 0
	}

	names := fs.Args()
	if len(names) == 0 {
		// No operands: print every environment variable as NAME=VALUE.
		for _, e := range os.Environ() {
			_, _ = fmt.Fprintf(stdio.Out, "%s%c", e, end)
		}
		return nil
	}

	// With operands: print each named variable's value, one per line. If any
	// requested variable is unset, the exit status is 1 (but set ones still
	// print).
	missing := false
	for _, name := range names {
		value, ok := os.LookupEnv(name)
		if !ok {
			missing = true
			continue
		}
		_, _ = fmt.Fprintf(stdio.Out, "%s%c", value, end)
	}
	if missing {
		return command.SilentFailure()
	}
	return nil
}
