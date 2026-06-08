// Package posixer implements the posixer applet: report whether the
// POSIX-defined utilities are installed on the system. It is a clean-room port
// of the maintainer's archived nao1215/posixer.
package posixer

import (
	"context"
	"fmt"
	"os/exec"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the posixer applet.
type Command struct{}

// New returns a posixer command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "posixer" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report which POSIX utilities are installed" }

// utility is one POSIX-defined command.
type utility struct {
	name string
	kind string // "required" or "optional"
}

// lookPath resolves a command to its absolute path; tests replace it.
var lookPath = exec.LookPath

// Run executes posixer.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[check]", stdio.Err)

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	// "posixer" and "posixer check" behave the same; any other operand is an
	// error so the usage stays discoverable.
	if rest := fs.Args(); len(rest) > 0 && rest[0] != "check" {
		return command.Failuref("unknown subcommand %q", rest[0])
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tTYPE\tINSTALLED\tPATH")
	for _, u := range posixUtilities {
		path, err := lookPath(u.name)
		installed, shown := "no", "-"
		if err == nil {
			installed, shown = "yes", path
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", u.name, u.kind, installed, shown)
	}
	if err := tw.Flush(); err != nil {
		return command.Failure(err)
	}
	return nil
}

// posixUtilities is an embedded subset of the POSIX.1 utility set. The list is
// representative rather than exhaustive; "required" marks commands a conforming
// system must provide, "optional" marks those in optional option groups.
var posixUtilities = []utility{
	{"awk", "required"}, {"basename", "required"}, {"cat", "required"},
	{"cd", "required"}, {"chgrp", "required"}, {"chmod", "required"},
	{"chown", "required"}, {"cmp", "required"}, {"cp", "required"},
	{"cut", "required"}, {"date", "required"}, {"dd", "required"},
	{"diff", "required"}, {"dirname", "required"}, {"echo", "required"},
	{"env", "required"}, {"expr", "required"}, {"false", "required"},
	{"find", "required"}, {"grep", "required"}, {"head", "required"},
	{"id", "required"}, {"kill", "required"}, {"ln", "required"},
	{"ls", "required"}, {"mkdir", "required"}, {"mv", "required"},
	{"od", "required"}, {"paste", "required"}, {"printf", "required"},
	{"pwd", "required"}, {"rm", "required"}, {"rmdir", "required"},
	{"sed", "required"}, {"sh", "required"}, {"sleep", "required"},
	{"sort", "required"}, {"tail", "required"}, {"tee", "required"},
	{"test", "required"}, {"touch", "required"}, {"tr", "required"},
	{"true", "required"}, {"uname", "required"}, {"uniq", "required"},
	{"wc", "required"}, {"xargs", "required"},
	{"bc", "optional"}, {"cron", "optional"}, {"ed", "optional"},
	{"mailx", "optional"}, {"vi", "optional"},
}
