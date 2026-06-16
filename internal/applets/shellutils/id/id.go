// Package id implements the id applet: print real and effective user and
// group IDs.
package id

import (
	"context"
	"fmt"
	"os/user"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the id applet.
type Command struct{}

// New returns an id command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "id" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print User ID and Group ID" }

// Run executes id.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... [USER]", stdio.Err).WithHelp(command.Help{
		Description: "Print user and group information for the current user, or for USER when " +
			"one is given. With no options, print the real and effective user ID, group ID, " +
			"and the list of supplementary groups.",
		Examples: []command.Example{
			{Command: "id", Explain: "Print all ID information for the current user."},
			{Command: "id -u", Explain: "Print only the effective user ID."},
			{Command: "id -un", Explain: "Print only the effective user name."},
			{Command: "id root", Explain: "Print ID information for the user root."},
		},
		ExitStatus: "0  the requested information was printed.\n1  an option was misused or the user could not be resolved.",
	})
	userOnly := fs.BoolP("user", "u", false, "print only the effective user ID")
	groupOnly := fs.BoolP("group", "g", false, "print only the effective group ID")
	allGroups := fs.BoolP("groups", "G", false, "print all group IDs")
	name := fs.BoolP("name", "n", false, "print a name instead of a number, for -ugG")
	real := fs.BoolP("real", "r", false, "print the real ID instead of the effective ID, with -ugG")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// At most one of -u, -g, -G may be specified at a time.
	if boolCount(*userOnly, *groupOnly, *allGroups) > 1 {
		_, _ = fmt.Fprintln(stdio.Err, "id: cannot print \"only\" of more than one choice")
		return command.SilentFailure()
	}
	// -n and -r are only meaningful together with -u, -g or -G.
	if (*name || *real) && !*userOnly && !*groupOnly && !*allGroups {
		_, _ = fmt.Fprintln(stdio.Err, "id: cannot print only names or real IDs in default format")
		return command.SilentFailure()
	}

	u, err := resolveUser(stdio, fs.Args())
	if err != nil {
		return err
	}

	switch {
	case *userOnly:
		return dumpUID(stdio, u, *name, *real)
	case *groupOnly:
		return dumpGID(stdio, u, *name, *real)
	case *allGroups:
		return dumpGroups(stdio, u, *name)
	default:
		return dumpAll(stdio, u)
	}
}

// resolveUser returns the current user, or the user named by the single
// operand when one is given.
func resolveUser(stdio command.IO, operands []string) (*user.User, error) {
	if len(operands) == 0 {
		u, err := user.Current()
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "id: %v\n", err)
			return nil, command.SilentFailure()
		}
		return u, nil
	}
	if len(operands) > 1 {
		_, _ = fmt.Fprintf(stdio.Err, "id: extra operand '%s'\n", operands[1])
		return nil, command.SilentFailure()
	}
	u, err := user.Lookup(operands[0])
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "id: '%s': no such user\n", operands[0])
		return nil, command.SilentFailure()
	}
	return u, nil
}

// dumpUID prints the user's ID (real ignored: real and effective are the same
// for a looked-up user) or, with showName, the user name.
func dumpUID(stdio command.IO, u *user.User, showName, _ bool) error {
	if showName {
		_, _ = fmt.Fprintln(stdio.Out, u.Username)
		return nil
	}
	_, _ = fmt.Fprintln(stdio.Out, u.Uid)
	return nil
}

// dumpGID prints the user's primary group ID or, with showName, its name.
func dumpGID(stdio command.IO, u *user.User, showName, _ bool) error {
	if showName {
		g, err := user.LookupGroupId(u.Gid)
		if err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "id: %v\n", err)
			return command.SilentFailure()
		}
		_, _ = fmt.Fprintln(stdio.Out, g.Name)
		return nil
	}
	_, _ = fmt.Fprintln(stdio.Out, u.Gid)
	return nil
}

// dumpGroups prints all of the user's group IDs or, with showName, their names.
func dumpGroups(stdio command.IO, u *user.User, showName bool) error {
	groups, err := lookupGroups(u)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "id: %v\n", err)
		return command.SilentFailure()
	}
	parts := make([]string, 0, len(groups))
	for _, g := range groups {
		if showName {
			parts = append(parts, g.Name)
		} else {
			parts = append(parts, g.Gid)
		}
	}
	_, _ = fmt.Fprintln(stdio.Out, strings.Join(parts, " "))
	return nil
}

// dumpAll prints the default GNU id line:
// uid=N(name) gid=N(name) groups=N(name),...
func dumpAll(stdio command.IO, u *user.User) error {
	primary, err := user.LookupGroupId(u.Gid)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "id: %v\n", err)
		return command.SilentFailure()
	}
	groups, err := lookupGroups(u)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "id: %v\n", err)
		return command.SilentFailure()
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "uid=%s(%s) gid=%s(%s) groups=", u.Uid, u.Username, u.Gid, primary.Name)
	parts := make([]string, 0, len(groups))
	for _, g := range groups {
		parts = append(parts, g.Gid+"("+g.Name+")")
	}
	b.WriteString(strings.Join(parts, ","))
	_, _ = fmt.Fprintln(stdio.Out, b.String())
	return nil
}

// lookupGroups returns the user's supplementary and primary groups.
func lookupGroups(u *user.User) ([]user.Group, error) {
	gids, err := u.GroupIds()
	if err != nil {
		return nil, err
	}
	groups := make([]user.Group, 0, len(gids))
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			return nil, err
		}
		groups = append(groups, *g)
	}
	return groups, nil
}

// boolCount returns how many of the given booleans are true.
func boolCount(bs ...bool) int {
	n := 0
	for _, b := range bs {
		if b {
			n++
		}
	}
	return n
}
