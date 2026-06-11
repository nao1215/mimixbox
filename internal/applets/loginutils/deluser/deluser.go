// Package deluser implements the deluser applet: remove a user account from
// /etc/passwd and /etc/shadow, and from group memberships.
package deluser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the deluser applet.
type Command struct{}

// New returns a deluser command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "deluser" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove a user account" }

// The account databases; tests point these at fixtures.
var (
	passwdPath = "/etc/passwd"
	shadowPath = "/etc/shadow"
	groupPath  = "/etc/group"
)

// Run executes deluser.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "USER", stdio.Err).WithHelp(command.Help{
		Description: "Remove the user account USER from /etc/passwd and /etc/shadow, and remove it from " +
			"the member list of every group in /etc/group. It is an error if the user does not exist. " +
			"Writing the real databases requires privilege.",
		Examples: []command.Example{
			{Command: "deluser alice", Explain: "Remove the user alice."},
		},
		ExitStatus: "0  the user was removed.\n1  no such user, or a database is unwritable.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a user name is required")
	}
	name := rest[0]

	passwd, err := readLines(passwdPath)
	if err != nil {
		return command.Failuref("cannot read %s: %v", passwdPath, err)
	}
	kept, removed := removeByName(passwd, name)
	if !removed {
		return command.Failuref("user %q does not exist", name)
	}
	if err := writeFile(passwdPath, kept, 0o644); err != nil {
		return command.Failuref("cannot write %s: %v", passwdPath, err)
	}

	shadow, err := readLines(shadowPath)
	if err != nil {
		return command.Failuref("cannot read %s: %v", shadowPath, err)
	}
	shadowKept, _ := removeByName(shadow, name)
	if err := writeFile(shadowPath, shadowKept, 0o600); err != nil {
		return command.Failuref("cannot write %s: %v", shadowPath, err)
	}

	groups, err := readLines(groupPath)
	if err != nil {
		return command.Failuref("cannot read %s: %v", groupPath, err)
	}
	if err := writeFile(groupPath, removeFromMembers(groups, name), 0o644); err != nil {
		return command.Failuref("cannot write %s: %v", groupPath, err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "deluser: user %q removed\n", name)
	return nil
}

// removeByName drops the line whose first colon-field equals name.
func removeByName(lines []string, name string) (kept []string, removed bool) {
	for _, line := range lines {
		if f := strings.Split(line, ":"); len(f) > 0 && f[0] == name {
			removed = true
			continue
		}
		kept = append(kept, line)
	}
	return kept, removed
}

// removeFromMembers drops name from the comma-separated member list (field 4)
// of every group line.
func removeFromMembers(lines []string, name string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 4 || fields[3] == "" {
			out = append(out, line)
			continue
		}
		var members []string
		for _, m := range strings.Split(fields[3], ",") {
			if m != name && m != "" {
				members = append(members, m)
			}
		}
		fields[3] = strings.Join(members, ",")
		out = append(out, strings.Join(fields, ":"))
	}
	return out
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // well-known account database path
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// writeFile atomically replaces path with the given lines.
func writeFile(path string, lines []string, mode os.FileMode) error {
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".deluser-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
