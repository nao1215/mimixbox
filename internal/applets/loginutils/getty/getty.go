// Package getty implements the getty applet: print the login banner, read a
// username, and hand off to login.
package getty

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the getty applet.
type Command struct{}

// New returns a getty command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "getty" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Prompt for a username and run login" }

// Injected so the banner source and the login hand-off are testable.
var (
	issuePath   = "/etc/issue"
	loginExecFn = func(stdio command.IO, username string) error {
		cmd := exec.Command("login", username) //nolint:gosec // handing off to login is the point
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		return cmd.Run()
	}
)

// Run executes getty.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[BAUD] TTY [TERM]", stdio.Err).WithHelp(command.Help{
		Description: "Open a terminal session: print the contents of /etc/issue (if present) and a " +
			"'login: ' prompt, read a username, and hand off to the login program for that user. The " +
			"BAUD, TTY, and TERM operands are accepted for compatibility; this build prompts on the " +
			"current terminal rather than opening the named device.",
		Examples: []command.Example{
			{Command: "getty 38400 tty1", Explain: "Prompt on tty1 and run login."},
		},
		ExitStatus: "login's exit status, or 1 if no TTY was given or no username was entered.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if len(fs.Args()) == 0 {
		return command.Failuref("a TTY argument is required")
	}

	if issue, err := os.ReadFile(issuePath); err == nil {
		_, _ = stdio.Out.Write(issue)
	}
	_, _ = fmt.Fprint(stdio.Out, "login: ")

	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		return command.Failuref("no username entered")
	}
	username := strings.TrimSpace(sc.Text())
	if username == "" {
		return command.Failuref("no username entered")
	}

	if err := loginExecFn(stdio, username); err != nil {
		var ee *exec.ExitError
		if e, isExit := err.(*exec.ExitError); isExit {
			ee = e
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("login: %v", err)
	}
	return nil
}
