// Package selinux implements the read-only SELinux query applets (getenforce,
// selinuxenabled, sestatus, getsebool, matchpathcon) plus deterministic,
// documented stubs for the privileged mutating applets (setenforce, setsebool,
// chcon, runcon, restorecon, setfiles, load_policy).
//
// All state is read through an injectable backend so the commands can be tested
// hermetically without touching the host's /sys/fs/selinux mount.
package selinux

import (
	"context"

	"github.com/nao1215/mimixbox/internal/command"
)

// command name constants for the multi-applet package.
const (
	cmdGetenforce     = "getenforce"
	cmdSelinuxenabled = "selinuxenabled"
	cmdSestatus       = "sestatus"
	cmdGetsebool      = "getsebool"
	cmdMatchpathcon   = "matchpathcon"
	cmdSetenforce     = "setenforce"
	cmdSetsebool      = "setsebool"
	cmdChcon          = "chcon"
	cmdRuncon         = "runcon"
	cmdRestorecon     = "restorecon"
	cmdSetfiles       = "setfiles"
	cmdLoadPolicy     = "load_policy"
)

// Command is one SELinux applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Constructors for each applet.

// NewGetenforce returns the getenforce applet.
func NewGetenforce() *Command { return &Command{name: cmdGetenforce} }

// NewSelinuxenabled returns the selinuxenabled applet.
func NewSelinuxenabled() *Command { return &Command{name: cmdSelinuxenabled} }

// NewSestatus returns the sestatus applet.
func NewSestatus() *Command { return &Command{name: cmdSestatus} }

// NewGetsebool returns the getsebool applet.
func NewGetsebool() *Command { return &Command{name: cmdGetsebool} }

// NewMatchpathcon returns the matchpathcon applet.
func NewMatchpathcon() *Command { return &Command{name: cmdMatchpathcon} }

// NewSetenforce returns the setenforce applet.
func NewSetenforce() *Command { return &Command{name: cmdSetenforce} }

// NewSetsebool returns the setsebool applet.
func NewSetsebool() *Command { return &Command{name: cmdSetsebool} }

// NewChcon returns the chcon applet.
func NewChcon() *Command { return &Command{name: cmdChcon} }

// NewRuncon returns the runcon applet.
func NewRuncon() *Command { return &Command{name: cmdRuncon} }

// NewRestorecon returns the restorecon applet.
func NewRestorecon() *Command { return &Command{name: cmdRestorecon} }

// NewSetfiles returns the setfiles applet.
func NewSetfiles() *Command { return &Command{name: cmdSetfiles} }

// NewLoadPolicy returns the load_policy applet.
func NewLoadPolicy() *Command { return &Command{name: cmdLoadPolicy} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case cmdGetenforce:
		return "Print the current SELinux enforcing mode"
	case cmdSelinuxenabled:
		return "Exit 0 if SELinux is enabled, 1 otherwise"
	case cmdSestatus:
		return "Show the SELinux status summary"
	case cmdGetsebool:
		return "Show the state of SELinux booleans"
	case cmdMatchpathcon:
		return "Show the default file context for a path"
	case cmdSetenforce:
		return "Set the SELinux enforcing mode (privileged)"
	case cmdSetsebool:
		return "Set the state of an SELinux boolean (privileged)"
	case cmdChcon:
		return "Change the SELinux security context of files (privileged)"
	case cmdRuncon:
		return "Run a program in a different SELinux context (privileged)"
	case cmdRestorecon:
		return "Restore default SELinux contexts on files (privileged)"
	case cmdSetfiles:
		return "Set file SELinux contexts from a spec file (privileged)"
	case cmdLoadPolicy:
		return "Load a new SELinux policy into the kernel (privileged)"
	}
	return "SELinux utility"
}

// Run dispatches to the per-command implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case cmdGetenforce:
		return c.runGetenforce(stdio, args)
	case cmdSelinuxenabled:
		return c.runSelinuxenabled(stdio, args)
	case cmdSestatus:
		return c.runSestatus(stdio, args)
	case cmdGetsebool:
		return c.runGetsebool(stdio, args)
	case cmdMatchpathcon:
		return c.runMatchpathcon(stdio, args)
	default:
		return c.runPrivileged(stdio, args)
	}
}
