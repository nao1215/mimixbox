// Package unshare implements the unshare applet: run a program with some
// namespaces unshared from the parent.
package unshare

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the unshare applet.
type Command struct{}

// New returns an unshare command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "unshare" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program with unshared namespaces" }

// unshareFn is indirected so the namespace flags can be tested without the
// privilege most namespaces require.
var unshareFn = func(flags int) error { return unix.Unshare(flags) }

// namespace describes one --flag and the clone bit it sets.
var namespaces = []struct {
	long, short string
	bit         int
	help        string
}{
	{"mount", "m", unix.CLONE_NEWNS, "unshare the mount namespace"},
	{"uts", "u", unix.CLONE_NEWUTS, "unshare the UTS (hostname) namespace"},
	{"ipc", "i", unix.CLONE_NEWIPC, "unshare the System V IPC namespace"},
	{"net", "n", unix.CLONE_NEWNET, "unshare the network namespace"},
	{"pid", "p", unix.CLONE_NEWPID, "unshare the PID namespace"},
	{"user", "U", unix.CLONE_NEWUSER, "unshare the user namespace"},
	{"cgroup", "C", unix.CLONE_NEWCGROUP, "unshare the cgroup namespace"},
}

// Run executes unshare.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[namespaces] [COMMAND [ARG...]]", stdio.Err).WithHelp(command.Help{
		Description: "Unshare the selected namespaces from the parent process, then run COMMAND (or a " +
			"shell if none is given) in the new namespaces. Most namespaces require privilege; the " +
			"user namespace (-U) can usually be unshared unprivileged.",
		Examples: []command.Example{
			{Command: "unshare -u hostname newname", Explain: "Run with a private UTS namespace."},
			{Command: "unshare --net ip addr", Explain: "Run with a fresh network namespace."},
		},
		ExitStatus: "the command's exit status, or 1 if a namespace could not be unshared.",
	})
	flags := map[string]*bool{}
	for _, ns := range namespaces {
		flags[ns.long] = fs.BoolP(ns.long, ns.short, false, ns.help)
	}
	// Stop at the first operand so the command's own flags (e.g. sh -c) are
	// passed through rather than parsed as unshare options.
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	var clone int
	for _, ns := range namespaces {
		if *flags[ns.long] {
			clone |= ns.bit
		}
	}
	if clone == 0 {
		return command.Failuref("at least one namespace flag is required")
	}

	if err := unshareFn(clone); err != nil {
		return command.Failuref("unshare failed: %v", err)
	}

	name, cmdArgs := shellOr(fs.Args())
	cmd := exec.Command(name, cmdArgs...) //nolint:gosec // user-specified command, as unshare is designed to run
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

// shellOr returns the command to run: the operands, or the login shell.
func shellOr(operands []string) (string, []string) {
	if len(operands) > 0 {
		return operands[0], operands[1:]
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	return shell, nil
}
