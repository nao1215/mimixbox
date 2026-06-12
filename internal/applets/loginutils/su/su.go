// Package su implements the su applet: run a shell or command as another user
// after authenticating.
package su

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

// Command is the su applet.
type Command struct{}

// New returns a su command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "su" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a shell as another user" }

// account is the resolved target user.
type account struct {
	name     string
	uid, gid int
	home     string
	shell    string
}

// Injected so the password database, the privilege check, the authentication
// backend, and the privileged exec are all testable.
var (
	passwdPath = "/etc/passwd"
	isRootFn   = func() bool { return os.Geteuid() == 0 }
	authFn     = auth.Authenticate
	runFn = func(stdio command.IO, acc account, argv []string) error {
		cmd := exec.Command(acc.shell) //nolint:gosec // running the user's shell is the point
		cmd.Args = argv
		cmd.Dir = acc.home
		cmd.Env = append(os.Environ(), "HOME="+acc.home, "USER="+acc.name, "SHELL="+acc.shell)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = stdio.In, stdio.Out, stdio.Err
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{Uid: uint32(acc.uid), Gid: uint32(acc.gid)},
		}
		return cmd.Run()
	}
)

// Run executes su.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-l] [-s SHELL] [-c COMMAND] [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Run a shell as USER (root by default), after authenticating unless already root. " +
			"-c runs a single COMMAND with the shell instead of an interactive session; -s overrides " +
			"the shell; -l (or a bare '-' operand) starts a login shell. The password is read from " +
			"standard input. Switching to another user requires privilege.",
		Examples: []command.Example{
			{Command: "su -c 'id' alice", Explain: "Run id as alice."},
			{Command: "su - root", Explain: "Start a root login shell."},
		},
		ExitStatus: "the command's exit status, or 1 on authentication failure or a bad user.",
	})
	login := fs.BoolP("login", "l", false, "start a login shell")
	shellOverride := fs.StringP("shell", "s", "", "shell to run instead of the user's")
	cmdStr := fs.StringP("command", "c", "", "run COMMAND with the shell")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	// A bare "-" operand means a login shell.
	var rest []string
	for _, op := range operands {
		if op == "-" {
			*login = true
			continue
		}
		rest = append(rest, op)
	}
	target := "root"
	if len(rest) > 0 {
		target = rest[0]
	}

	acc, err := resolve(target)
	if err != nil {
		return command.Failuref("%v", err)
	}
	if *shellOverride != "" {
		acc.shell = *shellOverride
	}
	if acc.shell == "" {
		acc.shell = "/bin/sh"
	}

	if !isRootFn() {
		ok, err := authenticate(stdio, target)
		if err != nil {
			return command.Failuref("%v", err)
		}
		if !ok {
			return command.Failuref("authentication failure")
		}
	}

	if err := runFn(stdio, acc, buildArgv(acc.shell, *login, *cmdStr)); err != nil {
		var ee *exec.ExitError
		if asExit(err, &ee) {
			return &command.ExitError{Code: ee.ExitCode()}
		}
		return command.Failuref("%v", err)
	}
	return nil
}

// buildArgv constructs the shell's argv, honoring the login-shell convention
// (argv[0] prefixed with '-') and an optional -c command.
func buildArgv(shell string, login bool, cmdStr string) []string {
	argv0 := filepath.Base(shell)
	if login {
		argv0 = "-" + argv0
	}
	argv := []string{argv0}
	if cmdStr != "" {
		argv = append(argv, "-c", cmdStr)
	}
	return argv
}

// resolve looks up the target user in the passwd database.
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

// authenticate reads a password from stdin and verifies it for the user.
func authenticate(stdio command.IO, user string) (bool, error) {
	sc := bufio.NewScanner(stdio.In)
	if !sc.Scan() {
		return false, fmt.Errorf("no password provided")
	}
	return authFn(user, sc.Text())
}

// asExit reports whether err is an *exec.ExitError and stores it.
func asExit(err error, target **exec.ExitError) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		*target = ee
		return true
	}
	return false
}
