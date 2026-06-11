// Package crontab implements the crontab applet: install, list, or remove a
// user's crontab in the cron spool directory.
package crontab

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the crontab applet.
type Command struct{}

// New returns a crontab command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "crontab" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Maintain a user's crontab" }

// Injected so the spool directory and the current user are testable.
var (
	spoolDir      = "/var/spool/cron/crontabs"
	currentUserFn = func() (string, error) {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		return u.Username, nil
	}
)

// Run executes crontab.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-u USER] {-l | -r | [FILE]}", stdio.Err).WithHelp(command.Help{
		Description: "Maintain a user's crontab in the cron spool directory. With -l the crontab is " +
			"printed; with -r it is removed; with a FILE (or '-' / no operand, meaning standard input) " +
			"the crontab is replaced with that content. -u operates on another user's crontab and " +
			"requires privilege. Interactive editing (-e) is not implemented.",
		Examples: []command.Example{
			{Command: "crontab -l", Explain: "Print the current crontab."},
			{Command: "crontab mycron.txt", Explain: "Install mycron.txt as the crontab."},
			{Command: "crontab -r", Explain: "Remove the crontab."},
		},
		ExitStatus: "0  the operation succeeded.\n1  no crontab, an unsupported mode, or an I/O error.",
	})
	list := fs.BoolP("list", "l", false, "print the crontab")
	remove := fs.BoolP("remove", "r", false, "remove the crontab")
	edit := fs.BoolP("edit", "e", false, "edit the crontab (not supported)")
	userName := fs.StringP("user", "u", "", "operate on this user's crontab")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	target := *userName
	if target == "" {
		target, err = currentUserFn()
		if err != nil {
			return command.Failuref("cannot determine the current user: %v", err)
		}
	}
	path := filepath.Join(spoolDir, target)

	switch {
	case *edit:
		return command.Failuref("interactive editing (-e) is not supported; install a file instead")
	case *list:
		return listCrontab(stdio, path)
	case *remove:
		return removeCrontab(path)
	default:
		return installCrontab(stdio, path, fs.Args())
	}
}

func listCrontab(stdio command.IO, path string) error {
	data, err := os.ReadFile(path) //nolint:gosec // spool path
	if err != nil {
		return command.Failuref("no crontab for this user")
	}
	_, _ = stdio.Out.Write(data)
	return nil
}

func removeCrontab(path string) error {
	if err := os.Remove(path); err != nil {
		return command.Failuref("no crontab to remove")
	}
	return nil
}

func installCrontab(stdio command.IO, path string, operands []string) error {
	var data []byte
	var err error
	if len(operands) == 0 || operands[0] == "-" {
		data, err = io.ReadAll(stdio.In)
	} else {
		data, err = os.ReadFile(operands[0]) //nolint:gosec // user-named crontab file
	}
	if err != nil {
		return command.Failuref("cannot read the new crontab: %v", err)
	}

	if err := os.MkdirAll(spoolDir, 0o755); err != nil {
		return command.Failuref("cannot create the spool directory: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return command.Failuref("cannot install the crontab: %v", err)
	}
	_, _ = fmt.Fprintln(stdio.Err, "crontab: installing new crontab")
	return nil
}
