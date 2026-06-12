// Package login implements the login applet: authenticate a user and start
// their login shell.
package login

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/nao1215/mimixbox/internal/auth"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the login applet.
type Command struct{}

// New returns a login command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "login" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Authenticate a user and start their shell" }

// account is the resolved user.
type account struct {
	name     string
	uid, gid int
	home     string
	shell    string
}

// Injected so the database, the auth backend, and the privileged exec are testable.
var (
	passwdPath = "/etc/passwd"
	authFn     = auth.Authenticate
	runFn      = func(stdio command.IO, acc account) error {
		shell := acc.shell
		if shell == "" {
			shell = "/bin/sh"
		}
		cmd := exec.Command(shell) //nolint:gosec // starting the user's login shell is the point
		cmd.Args = []string{"-" + filepath.Base(shell)}
		cmd.Dir = acc.home
		cmd.Env = append(os.Environ(), "HOME="+acc.home, "USER="+acc.name, "SHELL="+shell)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: uint32(acc.uid), Gid: uint32(acc.gid)},
		}
		return cmd.Run()
	}
)

// Run executes login.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-f USER] [-p] [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Authenticate a user and start their login shell with their credentials. The user " +
			"is taken from -f (pre-authenticated, no password asked), from the USER operand (then the " +
			"password is read from standard input), or — if neither is given — the username is read as " +
			"the first line of standard input and the password as the second. -p preserves the " +
			"environment. Logging in another user requires privilege.",
		Examples: []command.Example{
			{Command: "login alice", Explain: "Log in alice, reading the password from stdin."},
			{Command: "login -f root", Explain: "Start root's shell without a password."},
		},
		ExitStatus: "the shell's exit status, or 1 on a bad login.",
	})
	noAuth := fs.StringP("force", "f", "", "log in this user without authenticating")
	_ = fs.BoolP("preserve", "p", false, "preserve the environment")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	user, password, err := credentials(stdio, *noAuth, fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}

	acc, err := resolve(user)
	if err != nil {
		return command.Failuref("%v", err)
	}

	if *noAuth == "" {
		ok, err := authFn(user, password)
		if err != nil {
			return command.Failuref("%v", err)
		}
		if !ok {
			return command.Failuref("Login incorrect")
		}
	}

	if err := runFn(stdio, acc); err != nil {
		var ee *exec.ExitError
		if e, isExit := err.(*exec.ExitError); isExit {
			ee = e
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// credentials determines the user name and password to use.
func credentials(stdio command.IO, noAuth string, operands []string) (user, password string, err error) {
	if noAuth != "" {
		return noAuth, "", nil
	}
	sc := bufio.NewScanner(stdio.In)
	if len(operands) > 0 {
		user = operands[0]
	} else {
		_, _ = fmt.Fprint(stdio.Err, "login: ")
		if !sc.Scan() {
			return "", "", fmt.Errorf("no username provided")
		}
		user = strings.TrimSpace(sc.Text())
	}
	_, _ = fmt.Fprint(stdio.Err, "Password: ")
	if !sc.Scan() {
		return "", "", fmt.Errorf("no password provided")
	}
	return user, sc.Text(), nil
}

// resolve looks up the user in the passwd database.
func resolve(name string) (account, error) {
	data, err := os.ReadFile(passwdPath) //nolint:gosec // well-known passwd path
	if err != nil {
		return account{}, fmt.Errorf("cannot read %s: %v", passwdPath, err)
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, ":")
		if len(f) < 7 || f[0] != name {
			continue
		}
		uid, err1 := strconv.Atoi(f[2])
		gid, err2 := strconv.Atoi(f[3])
		if err1 != nil || err2 != nil {
			return account{}, fmt.Errorf("user %q has an invalid uid/gid", name)
		}
		return account{name: name, uid: uid, gid: gid, home: f[5], shell: f[6]}, nil
	}
	return account{}, fmt.Errorf("unknown user: %s", name)
}
