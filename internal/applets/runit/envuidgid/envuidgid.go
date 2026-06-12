// Package envuidgid implements the envuidgid applet: run a program with $UID and
// $GID set from a given account, without changing privilege.
package envuidgid

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the envuidgid applet.
type Command struct{}

// New returns an envuidgid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "envuidgid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program with $UID/$GID from a user" }

// passwdPath is the password database; tests point it at a fixture.
var passwdPath = "/etc/passwd"

// Run executes envuidgid.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "USER PROG [ARG...]", stdio.Err).WithHelp(command.Help{
		Description: "Set the environment variables $UID and $GID to the numeric uid and gid of USER " +
			"(from /etc/passwd), then run PROG with its arguments. Unlike setuidgid it does not change " +
			"privilege; it only exports the ids, as the daemontools/runit envuidgid does.",
		Examples: []command.Example{
			{Command: "envuidgid nobody printenv UID GID", Explain: "Run printenv with nobody's ids."},
		},
		ExitStatus: "PROG's exit status, or 1 on a usage error or unknown user.",
	})
	// Stop at the first operand so PROG's own flags are not parsed by envuidgid.
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

	cmd := exec.CommandContext(ctx, rest[1], rest[2:]...) //nolint:gosec // running the user's program is the point
	cmd.Env = append(os.Environ(), "UID="+strconv.Itoa(uid), "GID="+strconv.Itoa(gid))
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// resolve returns the uid and gid of the named account.
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
