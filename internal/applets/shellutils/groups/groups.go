// Package groups implements the groups applet: print the groups a user
// belongs to. With no operand it prints the current user's groups; with one or
// more USERNAME operands it prints each user's groups.
package groups

import (
	"context"
	"fmt"
	"os/user"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the groups applet.
type Command struct{}

// New returns a groups command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "groups" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the groups to which USERNAME belongs" }

// Run executes groups.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[USERNAME]...", stdio.Err).WithHelp(command.Help{
		Description: "Print the names of the groups each USERNAME belongs to. With no operand, print the groups of " +
			"the current user.",
		Examples: []command.Example{
			{Command: "groups", Explain: "Print the groups the current user belongs to."},
			{Command: "groups root", Explain: "Print the groups the root user belongs to."},
			{Command: "groups alice bob", Explain: "Print the groups for several users, one per line."},
		},
		ExitStatus: "0  all named users were found.\n1  a named user could not be found.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	names := fs.Args()
	if len(names) == 0 {
		// No operand: print the current user's groups.
		u, uerr := user.Current()
		if uerr != nil {
			return command.Failuref("groups: cannot find current user: %v", uerr)
		}
		line, gerr := groupNames(u)
		if gerr != nil {
			return command.Failuref("groups: %v", gerr)
		}
		_, _ = fmt.Fprintln(stdio.Out, line)
		return nil
	}

	// One or more operands: print each user's groups.
	failed := false
	for _, name := range names {
		u, uerr := user.Lookup(name)
		if uerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "groups: '%s': no such user\n", name)
			failed = true
			continue
		}
		line, gerr := groupNames(u)
		if gerr != nil {
			_, _ = fmt.Fprintf(stdio.Err, "groups: '%s': %v\n", name, gerr)
			failed = true
			continue
		}
		// GNU prefixes each line with "user :" when an operand is given.
		if len(names) > 1 {
			_, _ = fmt.Fprintf(stdio.Out, "%s : %s\n", name, line)
		} else {
			_, _ = fmt.Fprintln(stdio.Out, line)
		}
	}
	if failed {
		return command.SilentFailure()
	}
	return nil
}

// groupNames resolves the group names u belongs to, space-separated.
func groupNames(u *user.User) (string, error) {
	gids, err := u.GroupIds()
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(gids))
	for _, gid := range gids {
		g, err := user.LookupGroupId(gid)
		if err != nil {
			// Fall back to the numeric id when the group has no name entry.
			names = append(names, gid)
			continue
		}
		names = append(names, g.Name)
	}
	return strings.Join(names, " "), nil
}
