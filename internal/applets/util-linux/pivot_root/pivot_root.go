// Package pivotroot implements the pivot_root applet: change the root
// filesystem of the current mount namespace.
package pivotroot

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the pivot_root applet.
type Command struct{}

// New returns a pivot_root command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "pivot_root" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Change the root filesystem" }

// pivotFn is indirected so the privileged syscall can be tested.
var pivotFn = func(newRoot, putOld string) error { return unix.PivotRoot(newRoot, putOld) }

// Run executes pivot_root.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NEW_ROOT PUT_OLD", stdio.Err).WithHelp(command.Help{
		Description: "Move the root filesystem of the current process to the directory PUT_OLD and make " +
			"NEW_ROOT the new root. Both must be directories, NEW_ROOT must be a mount point, and " +
			"PUT_OLD must be under NEW_ROOT. This changes the mount namespace and requires privilege.",
		Examples: []command.Example{
			{Command: "pivot_root /newroot /newroot/oldroot", Explain: "Pivot onto /newroot."},
		},
		ExitStatus: "0  the root was pivoted.\n1  the arguments were wrong or the syscall failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	if len(rest) != 2 {
		return command.Failuref("exactly two directories (NEW_ROOT and PUT_OLD) are required")
	}

	if err := pivotFn(rest[0], rest[1]); err != nil {
		return command.Failuref("%v", err)
	}
	return nil
}
