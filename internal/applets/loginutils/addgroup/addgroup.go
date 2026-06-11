// Package addgroup implements the addgroup applet: add a group to /etc/group.
package addgroup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the addgroup applet.
type Command struct{}

// New returns an addgroup command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "addgroup" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Add a group to /etc/group" }

// groupPath is the group database; tests point it at a fixture.
var groupPath = "/etc/group"

// The range of GIDs auto-assigned when --gid is not given.
const (
	firstGID = 1000
	lastGID  = 64999
)

// Run executes addgroup.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[--gid GID] GROUP", stdio.Err).WithHelp(command.Help{
		Description: "Add a new group named GROUP to /etc/group. With --gid the group is given that " +
			"numeric ID; otherwise the first free ID in the normal range is chosen. It is an error if " +
			"the group name or a requested GID already exists. Writing the real group database requires " +
			"privilege.",
		Examples: []command.Example{
			{Command: "addgroup developers", Explain: "Add a group with an auto-assigned GID."},
			{Command: "addgroup --gid 1500 staff", Explain: "Add a group with a specific GID."},
		},
		ExitStatus: "0  the group was added.\n1  the group or GID already exists, or the database is unwritable.",
	})
	gid := fs.Int("gid", -1, "numeric group ID (auto-assigned if unset)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return command.Failuref("a group name is required")
	}
	name := rest[0]

	lines, byName, byGID, err := readGroups()
	if err != nil {
		return command.Failuref("cannot read %s: %v", groupPath, err)
	}
	if byName[name] {
		return command.Failuref("group %q already exists", name)
	}

	newGID := *gid
	if fs.Changed("gid") {
		if byGID[newGID] {
			return command.Failuref("GID %d is already in use", newGID)
		}
	} else {
		newGID = nextFreeGID(byGID)
		if newGID < 0 {
			return command.Failuref("no free GID available in %d-%d", firstGID, lastGID)
		}
	}

	lines = append(lines, fmt.Sprintf("%s:x:%d:", name, newGID))
	if err := writeGroups(lines); err != nil {
		return command.Failuref("cannot write %s: %v", groupPath, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "addgroup: group %q (GID %d) added\n", name, newGID)
	return nil
}

// readGroups returns the group file's lines and indexes of used names and GIDs.
func readGroups() (lines []string, byName map[string]bool, byGID map[int]bool, err error) {
	byName = map[string]bool{}
	byGID = map[int]bool{}
	data, err := os.ReadFile(groupPath) //nolint:gosec // well-known group path
	if err != nil {
		return nil, nil, nil, err
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil, byName, byGID, nil
	}
	lines = strings.Split(trimmed, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}
		byName[fields[0]] = true
		if g, err := strconv.Atoi(fields[2]); err == nil {
			byGID[g] = true
		}
	}
	return lines, byName, byGID, nil
}

// nextFreeGID returns the lowest unused GID in the normal range, or -1.
func nextFreeGID(byGID map[int]bool) int {
	for g := firstGID; g <= lastGID; g++ {
		if !byGID[g] {
			return g
		}
	}
	return -1
}

// writeGroups atomically replaces the group file.
func writeGroups(lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	tmp, err := os.CreateTemp(filepath.Dir(groupPath), ".addgroup-*")
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
