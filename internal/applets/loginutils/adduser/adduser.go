// Package adduser implements the adduser applet: create a user account in
// /etc/passwd and /etc/shadow, with a primary group.
package adduser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the adduser applet.
type Command struct{}

// New returns an adduser command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "adduser" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Create a user account" }

// The account databases; tests point these at fixtures.
var (
	passwdPath = "/etc/passwd"
	shadowPath = "/etc/shadow"
	groupPath  = "/etc/group"
)

const (
	firstID = 1000
	lastID  = 64999
)

// Run executes adduser.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-u UID] [-h HOME] [-s SHELL] [-G GROUP] USER", stdio.Err).WithHelp(command.Help{
		Description: "Create the user account USER in /etc/passwd and /etc/shadow. -u sets the UID " +
			"(auto-assigned otherwise); -h sets the home directory (default /home/USER); -s sets the " +
			"login shell (default /bin/sh); -G makes an existing group the primary group, otherwise a " +
			"new group named USER is created. The account is created locked; set a password with " +
			"passwd or chpasswd. Writing the real databases requires privilege.",
		Examples: []command.Example{
			{Command: "adduser alice", Explain: "Create alice with a matching group."},
			{Command: "adduser -G staff -s /bin/bash bob", Explain: "Create bob in the staff group."},
		},
		ExitStatus: "0  the account was created.\n1  the user exists or a database is unwritable.",
	})
	uid := fs.IntP("uid", "u", -1, "numeric user ID (auto-assigned if unset)")
	home := fs.StringP("home", "h", "", "home directory (default /home/USER)")
	shell := fs.StringP("shell", "s", "/bin/sh", "login shell")
	group := fs.StringP("ingroup", "G", "", "existing group to use as the primary group")

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
	usedNames, usedUIDs := indexColumn(passwd, 0), indexNumeric(passwd, 2)
	if usedNames[name] {
		return command.Failuref("user %q already exists", name)
	}

	newUID := *uid
	if fs.Changed("uid") {
		if usedUIDs[newUID] {
			return command.Failuref("UID %d is already in use", newUID)
		}
	} else if newUID = nextFree(usedUIDs); newUID < 0 {
		return command.Failuref("no free UID available in %d-%d", firstID, lastID)
	}

	gid, err := resolveGroup(name, newUID, *group)
	if err != nil {
		return command.Failuref("%v", err)
	}

	homeDir := *home
	if homeDir == "" {
		homeDir = "/home/" + name
	}

	passwd = append(passwd, fmt.Sprintf("%s:x:%d:%d:%s:%s:%s", name, newUID, gid, name, homeDir, *shell))
	if err := writeFile(passwdPath, passwd, 0o644); err != nil {
		return command.Failuref("cannot write %s: %v", passwdPath, err)
	}

	shadow, err := readLines(shadowPath)
	if err != nil {
		return command.Failuref("cannot read %s: %v", shadowPath, err)
	}
	shadow = append(shadow, fmt.Sprintf("%s:!:19000:0:99999:7:::", name))
	if err := writeFile(shadowPath, shadow, 0o600); err != nil {
		return command.Failuref("cannot write %s: %v", shadowPath, err)
	}

	_, _ = fmt.Fprintf(stdio.Out, "adduser: user %q (UID %d, GID %d) created\n", name, newUID, gid)
	return nil
}

// resolveGroup returns the primary GID: the GID of an existing -G group, or a
// freshly created group named after the user (with GID = uid).
func resolveGroup(name string, uid int, ingroup string) (int, error) {
	groups, err := readLines(groupPath)
	if err != nil {
		return 0, fmt.Errorf("cannot read %s: %v", groupPath, err)
	}
	if ingroup != "" {
		for _, line := range groups {
			f := strings.Split(line, ":")
			if len(f) >= 3 && f[0] == ingroup {
				gid, err := strconv.Atoi(f[2])
				if err != nil {
					return 0, fmt.Errorf("group %q has an invalid GID", ingroup)
				}
				return gid, nil
			}
		}
		return 0, fmt.Errorf("group %q does not exist", ingroup)
	}
	// Create a new group named after the user.
	usedGIDs := indexNumeric(groups, 2)
	gid := uid
	if usedGIDs[gid] {
		if gid = nextFree(usedGIDs); gid < 0 {
			return 0, fmt.Errorf("no free GID available")
		}
	}
	groups = append(groups, fmt.Sprintf("%s:x:%d:", name, gid))
	if err := writeFile(groupPath, groups, 0o644); err != nil {
		return 0, fmt.Errorf("cannot write %s: %v", groupPath, err)
	}
	return gid, nil
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

func indexColumn(lines []string, col int) map[string]bool {
	out := map[string]bool{}
	for _, line := range lines {
		if f := strings.Split(line, ":"); len(f) > col {
			out[f[col]] = true
		}
	}
	return out
}

func indexNumeric(lines []string, col int) map[int]bool {
	out := map[int]bool{}
	for _, line := range lines {
		if f := strings.Split(line, ":"); len(f) > col {
			if n, err := strconv.Atoi(f[col]); err == nil {
				out[n] = true
			}
		}
	}
	return out
}

func nextFree(used map[int]bool) int {
	for id := firstID; id <= lastID; id++ {
		if !used[id] {
			return id
		}
	}
	return -1
}

// writeFile atomically replaces path with the given lines.
func writeFile(path string, lines []string, mode os.FileMode) error {
	content := strings.Join(lines, "\n") + "\n"
	tmp, err := os.CreateTemp(filepath.Dir(path), ".adduser-*")
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
