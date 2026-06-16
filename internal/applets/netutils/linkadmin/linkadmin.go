// Package linkadmin implements the link/tunnel administration applets whose
// write paths are not yet safe for CI: tc, iptunnel, nameif, and slattach. They
// share one shape: parse and validate arguments, then report a deterministic
// capability error explaining that the privileged operation is intentionally
// deferred. This honours the "never a silent no-op" rule. Read-only inspect
// subcommands ("show"/"list") are accepted and produce an explicit "no entries"
// notice rather than failing, since inspection is the safe first step.
//
// This file is the shared core: the Command shape, the flag/help wiring every
// applet reuses, and the minimal validation helpers (inspect dispatch and arity
// checks). Each applet's own surface — constructor, spec, examples, and its
// deferral/inspect behaviour — lives in its dedicated file (tc.go, iptunnel.go,
// nameif.go, slattach.go).
package linkadmin

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// spec is the static description of one applet: the strings shown in help and
// the applet list.
type spec struct {
	name     string
	synopsis string
	usage    string
	desc     string
}

// Command is one link/tunnel administration applet. Each constructor wires in
// the applet's spec, its example list, and the runner that implements its
// specific (inspect or deferral) behaviour.
type Command struct {
	spec     spec
	examples []command.Example
	// addFlags lets an applet register options before parsing (e.g. slattach's
	// -p); nil when the applet takes no options.
	addFlags func(fs *command.FlagSet)
	// run implements the applet's behaviour over its parsed operands.
	run func(stdio command.IO, operands []string) error
}

// Name returns the command name.
func (c *Command) Name() string { return c.spec.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return c.spec.synopsis }

// Run parses arguments with the shared flag/help wiring and dispatches to the
// applet's runner.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), c.spec.usage, stdio.Err).WithHelp(command.Help{
		Description: c.spec.desc,
		Examples:    c.examples,
		ExitStatus: "0  an inspect subcommand printed its (possibly empty) result.\n" +
			"1  a mutating/privileged operation was requested, or arguments were invalid.",
		Notes: []string{"Mutating/privileged operations are intentionally deferred and fail deterministically."},
	})
	if c.addFlags != nil {
		c.addFlags(fs)
	}
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	return c.run(stdio, fs.Args())
}

// inspectNotice prints the deterministic "no entries" line shared by the
// inspect-only applets (tc and iptunnel).
func inspectNotice(stdio command.IO, name string) error {
	fmt.Fprintf(stdio.Out, "%s: no entries (inspect-only slice; the live kernel is not queried)\n", name)
	return nil
}

// deferMutating reports the shared error for a mutating subcommand on an
// inspect-only applet.
func deferMutating(sub string) error {
	return command.Failuref(
		"%q is a mutating subcommand and is intentionally deferred; only show/list are available", sub)
}
