// Package nsenter implements the nsenter applet: run a program in the
// namespaces of another process.
package nsenter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the nsenter applet.
type Command struct{}

// New returns an nsenter command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nsenter" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Run a program in another process's namespaces" }

// setnsFn is indirected so entering namespaces can be tested without privilege.
// It receives the /proc/PID/ns/<type> path to enter.
var setnsFn = func(nsPath string) error {
	f, err := os.Open(nsPath) //nolint:gosec // /proc namespace path
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return unix.Setns(int(f.Fd()), 0)
}

// namespace maps a --flag to the /proc/PID/ns file it enters.
var namespaces = []struct {
	long, short, nsFile, help string
}{
	{"mount", "m", "mnt", "enter the mount namespace"},
	{"uts", "u", "uts", "enter the UTS (hostname) namespace"},
	{"ipc", "i", "ipc", "enter the System V IPC namespace"},
	{"net", "n", "net", "enter the network namespace"},
	{"pid", "p", "pid", "enter the PID namespace"},
	{"user", "U", "user", "enter the user namespace"},
	{"cgroup", "C", "cgroup", "enter the cgroup namespace"},
}

// Run executes nsenter.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-t PID [namespaces] [COMMAND [ARG...]]", stdio.Err).WithHelp(command.Help{
		Description: "Enter one or more namespaces of the target process given by -t PID, then run " +
			"COMMAND (or a shell) inside them. Select namespaces with -m/-u/-i/-n/-p/-U/-C. Entering " +
			"another process's namespaces requires privilege.",
		Examples: []command.Example{
			{Command: "nsenter -t 1234 -n ip addr", Explain: "Run ip in process 1234's network namespace."},
			{Command: "nsenter -t 1234 -m -u sh", Explain: "Enter its mount and UTS namespaces."},
		},
		ExitStatus: "the command's exit status, or 1 if a namespace could not be entered.",
	})
	target := fs.IntP("target", "t", 0, "PID of the process whose namespaces to enter")
	flags := map[string]*bool{}
	for _, ns := range namespaces {
		flags[ns.long] = fs.BoolP(ns.long, ns.short, false, ns.help)
	}
	fs.SetInterspersed(false)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if !fs.Changed("target") {
		return command.Failuref("a target PID (-t) is required")
	}

	var selected []string
	for _, ns := range namespaces {
		if *flags[ns.long] {
			selected = append(selected, ns.nsFile)
		}
	}
	if len(selected) == 0 {
		return command.Failuref("at least one namespace flag is required")
	}

	for _, nsFile := range selected {
		path := fmt.Sprintf("/proc/%d/ns/%s", *target, nsFile)
		if err := setnsFn(path); err != nil {
			return command.Failuref("cannot enter %s: %v", path, err)
		}
	}

	name, cmdArgs := shellOr(fs.Args())
	cmd := exec.Command(name, cmdArgs...) //nolint:gosec // user-specified command, as nsenter is designed to run
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
