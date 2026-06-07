// Package chroot implements the chroot applet: run a command or an interactive
// shell with a special root directory.
package chroot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the chroot applet.
type Command struct{}

// New returns a chroot command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "chroot" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Run command or interactive shell with special root directory"
}

// Run executes chroot. It changes the root directory to NEWROOT and runs
// COMMAND (defaulting to the shell) inside it. This requires root privileges;
// when it cannot change the root directory it prints a GNU-style error and
// returns command.SilentFailure().
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NEWROOT [COMMAND [ARG]...]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) == 0 {
		_, _ = fmt.Fprintln(stdio.Err, "chroot: missing operand")
		return command.SilentFailure()
	}

	newRoot := os.ExpandEnv(operands[0])
	if err := syscall.Chroot(newRoot); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: cannot change root directory to '%s': %s\n",
			newRoot, reason(err))
		return command.SilentFailure()
	}

	//----------------From here, in the prison-------------------
	if err := os.Chdir("/"); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: cannot change root directory to '%s': %s\n",
			newRoot, reason(err))
		return command.SilentFailure()
	}

	name, argv := decideExecCommand(operands[1:])
	// Reset the environment variable SHELL for the jail environment.
	if err := os.Setenv("SHELL", name); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: %v\n", err)
		return command.SilentFailure()
	}

	// TODO: Reset UID and GID.
	// "/etc/passwd (uid name resolution file)" and "/etc/group (gid name
	// resolution file)" may be different between the original environment and
	// the jail environment. So, reset uid and gid in the jail environment.

	cmd := exec.Command(name, argv...) //nolint:gosec // running a user-named command is the whole point
	cmd.Stdin = stdio.In
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "chroot: %v\n", err)
		return command.SilentFailure()
	}
	return nil
}

// decideExecCommand resolves the command to run inside the jail. extra are the
// operands after NEWROOT. When none are given, the command is the shell taken
// from $SHELL (falling back to /bin/sh) run interactively.
func decideExecCommand(extra []string) (name string, argv []string) {
	if len(extra) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		return shell, []string{"-i"}
	}
	return extra[0], extra[1:]
}

// reason maps a chroot/chdir failure to the GNU-style trailing message.
func reason(err error) string {
	if os.IsNotExist(err) {
		return "No such file or directory"
	}
	if os.IsPermission(err) {
		return "Operation not permitted"
	}
	return err.Error()
}
