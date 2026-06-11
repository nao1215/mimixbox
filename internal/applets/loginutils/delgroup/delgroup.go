// Package delgroup implements the delgroup applet: remove a group from
// /etc/group.
package delgroup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the delgroup applet.
type Command struct{}

// New returns a delgroup command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "delgroup" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Remove a group from /etc/group" }

// groupPath is the group database; tests point it at a fixture.
var groupPath = "/etc/group"

// Run executes delgroup.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "GROUP", stdio.Err).WithHelp(command.Help{
		Description: "Remove the group named GROUP from /etc/group. It is an error if no such group " +
			"exists. Writing the real group database requires privilege.",
		Examples: []command.Example{
			{Command: "delgroup developers", Explain: "Remove the developers group."},
		},
		ExitStatus: "0  the group was removed.\n1  no such group, or the database is unwritable.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a group name is required")
	}
	name := rest[0]

	data, err := os.ReadFile(groupPath) //nolint:gosec // well-known group path
	if err != nil {
		return command.Failuref("cannot read %s: %v", groupPath, err)
	}

	kept, removed := remove(string(data), name)
	if !removed {
		return command.Failuref("group %q does not exist", name)
	}
	if err := writeGroups(kept); err != nil {
		return command.Failuref("cannot write %s: %v", groupPath, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "delgroup: group %q removed\n", name)
	return nil
}

// remove returns the group file content with the named group's line dropped and
// whether a line was removed.
func remove(content, name string) (kept []string, removed bool) {
	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return nil, false
	}
	for _, line := range strings.Split(trimmed, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) > 0 && fields[0] == name {
			removed = true
			continue
		}
		kept = append(kept, line)
	}
	return kept, removed
}

// writeGroups atomically replaces the group file.
func writeGroups(lines []string) error {
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	tmp, err := os.CreateTemp(filepath.Dir(groupPath), ".delgroup-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, groupPath)
}
