// Package sulogin implements the sulogin applet: authenticate root and start a
// single-user root shell, as used in emergency/maintenance boots.
package sulogin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/nao1215/mimixbox/internal/auth"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the sulogin applet.
type Command struct{}

// New returns a sulogin command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "sulogin" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Single-user root login" }

// Injected so the authentication backend and the shell exec are testable.
var (
	authFn = auth.Authenticate
	runFn  = func(stdio command.IO, shell string) error {
		cmd := exec.Command(shell) //nolint:gosec // running the root shell is the point
		cmd.Args = []string{"-" + baseName(shell)}
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: 0, Gid: 0},
		}
		return cmd.Run()
	}
)

// Run executes sulogin.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t SECONDS] [TTY]", stdio.Err).WithHelp(command.Help{
		Description: "Authenticate the root user and, on success, start a single-user root shell — the " +
			"prompt shown in emergency or maintenance boots. The root password is read from standard " +
			"input. The shell is taken from $SHELL, or /bin/sh. Starting a root shell requires privilege.",
		Examples: []command.Example{
			{Command: "sulogin", Explain: "Prompt for the root password and start a root shell."},
		},
		ExitStatus: "the shell's exit status, or 1 on authentication failure.",
	})
	_ = fs.IntP("timeout", "t", 0, "seconds to wait for input (accepted for compatibility)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	_, _ = fmt.Fprint(stdio.Err, "Give root password for maintenance\n(or press Control-D to continue): ")
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		return command.Failuref("no password provided")
	}

	ok, err := authFn("root", sc.Text())
	if err != nil {
		return command.Failuref("%v", err)
	}
	if !ok {
		return command.Failuref("incorrect root password")
	}

	if err := runFn(stdio, shell()); err != nil {
		var ee *exec.ExitError
		if e, isExit := err.(*exec.ExitError); isExit {
			ee = e
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// shell returns the shell to start: $SHELL, or /bin/sh.
func shell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/sh"
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
