// Package modutils implements the kernel-module mutation applets insmod, rmmod,
// modprobe, and depmod. Each separates metadata parsing / plan generation (which
// is hermetic and testable) from the privileged kernel mutation (which requires
// CAP_SYS_MODULE). The privileged step is intentionally gated: it fails
// deterministically with a documented requirement rather than silently doing
// nothing.
//
// The shared module-name/file parsing and dependency planning lives in
// planner.go; each applet's CLI surface lives in its own file (insmod.go,
// rmmod.go, modprobe.go, depmod.go) and delegates to that backend.
package modutils

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

const (
	cmdInsmod   = "insmod"
	cmdRmmod    = "rmmod"
	cmdModprobe = "modprobe"
	cmdDepmod   = "depmod"
)

// Command is one module-mutation applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// NewInsmod returns the insmod applet.
func NewInsmod() *Command { return &Command{name: cmdInsmod} }

// NewRmmod returns the rmmod applet.
func NewRmmod() *Command { return &Command{name: cmdRmmod} }

// NewModprobe returns the modprobe applet.
func NewModprobe() *Command { return &Command{name: cmdModprobe} }

// NewDepmod returns the depmod applet.
func NewDepmod() *Command { return &Command{name: cmdDepmod} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case cmdInsmod:
		return "Validate and (privileged) insert a kernel module"
	case cmdRmmod:
		return "Validate and (privileged) remove a kernel module"
	case cmdModprobe:
		return "Resolve dependencies and (privileged) load a module"
	case cmdDepmod:
		return "Build the module dependency list"
	}
	return "Kernel module utility"
}

// Run dispatches to the per-applet implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case cmdInsmod:
		return c.runInsmod(stdio, args)
	case cmdRmmod:
		return c.runRmmod(stdio, args)
	case cmdModprobe:
		return c.runModprobe(stdio, args)
	case cmdDepmod:
		return c.runDepmod(stdio, args)
	}
	return command.Failuref("%s: unknown module applet", c.name)
}
