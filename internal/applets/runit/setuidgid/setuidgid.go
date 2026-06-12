// Package setuidgid implements the setuidgid applet: run a program as the uid
// and gid of a given account.
package setuidgid

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the setuidgid applet.
type Command struct{}

// New returns a setuidgid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "setuidgid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program as a user's uid/gid" }

// Injected so the password database and the privileged exec are testable.
var (
	passwdPath = "/etc/passwd"
	runFn      = func(ctx context.Context, stdio command.IO, uid, gid int, prog string, args []string) error {
		cmd := exec.CommandContext(ctx, prog, args...) //nolint:gosec // running the user's program is the point
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)},
		}
		return cmd.Run()
	}
)

// Run executes setuidgid.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "USER PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Set the real and effective uid and gid to those of USER (and the supplementary " +
			"group list to that user's primary group), then run PROG with its arguments. This is the " +
			"daemontools/runit setuidgid; dropping privilege requires running as root.",
		Examples: []command.Example{
			{Command: "setuidgid nobody mydaemon", Explain: "Run mydaemon as the nobody account."},
		},
		ExitStatus: "PROG's exit status, or 1 on a usage error or unknown user.",
	})
	// Stop at the first operand so PROG's own flags are not parsed by setuidgid.
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("a user and a program are required")
	}
	uid, gid, err := resolve(rest[0])
	if err != nil {
		return command.Failuref("%v", err)
	}

	if err := runFn(ctx, stdio, uid, gid, rest[1], rest[2:]); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// resolve returns the uid and gid of the named account from the passwd database.
func resolve(name string) (uid, gid int, err error) {
	data, err := os.ReadFile(passwdPath) //nolint:gosec // well-known passwd path
	if err != nil {
		return 0, 0, err
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, ":")
		if len(f) < 4 || f[0] != name {
			continue
		}
		u, err1 := strconv.Atoi(f[2])
		g, err2 := strconv.Atoi(f[3])
		if err1 != nil || err2 != nil {
			return 0, 0, errors.New("user " + name + " has an invalid uid/gid")
		}
		return u, g, nil
	}
	return 0, 0, errors.New("unknown user: " + name)
}
