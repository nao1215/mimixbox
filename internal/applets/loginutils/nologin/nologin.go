// Package nologin implements the nologin applet: politely refuse a login and
// exit non-zero, as the shell of an account that may not log in.
package nologin

import (
	"context"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the nologin applet.
type Command struct{}

// New returns a nologin command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nologin" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Refuse a login and exit non-zero" }

// messageFile holds the optional custom refusal message; tests override it.
var messageFile = "/etc/nologin.txt"

// defaultMessage is printed when no message file is present.
const defaultMessage = "This account is currently not available."

// Run executes nologin.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print a refusal message and exit non-zero, so it can serve as the login shell of " +
			"an account that must not log in. If /etc/nologin.txt exists its contents are shown; " +
			"otherwise a default message is printed. Any arguments (e.g. a shell '-c command') are " +
			"ignored, and no command is ever run.",
		Examples: []command.Example{
			{Command: "nologin", Explain: "Print the refusal message and exit 1."},
		},
		ExitStatus: "1  always: the login is refused.",
	})
	// nologin must ignore the arguments a calling shell passes (e.g. "-c command"),
	// so only consult the flag parser for --help/--version; otherwise refuse
	// without complaining about unknown flags.
	for _, a := range args {
		if a == "--help" || a == "-h" || a == "--version" {
			if proceed, _ := fs.Parse(stdio, args); !proceed {
				return nil
			}
			break
		}
	}

	_, _ = stdio.Out.Write(message())
	return command.SilentFailure()
}

// message returns the refusal text: the message file verbatim if present and
// non-empty, otherwise the default with a trailing newline.
func message() []byte {
	if data, err := os.ReadFile(messageFile); err == nil && len(data) > 0 {
		return data
	}
	return []byte(defaultMessage + "\n")
}
