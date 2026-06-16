// Package netctl implements the privileged, kernel-mutating networking applets
// brctl, ifenslave, tunctl, vconfig, zcip, and nbd-client.
//
// Each of these reconfigures kernel network state (bridges, bonds, TUN/TAP
// devices, VLANs, link-local addressing, network block devices) through
// privileged syscalls that are not available in this environment. Following the
// batch rule "never ship a silent no-op", each applet fully validates its
// arguments and serializes the requested action into a deterministic Plan, then
// fails with a documented capability/backend error describing exactly what it
// would have done. The argument parsing and plan serialization are pure and
// table-tested.
package netctl

import (
	"context"
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Plan is the validated, serialized action a netctl applet would perform.
type Plan struct {
	Tool   string   // applet name
	Action string   // sub-action (e.g. "addbr", "add-vlan")
	Args   []string // normalized operands
}

// String renders the plan as a stable, human- and test-readable line.
func (p Plan) String() string {
	if len(p.Args) == 0 {
		return fmt.Sprintf("%s %s", p.Tool, p.Action)
	}
	return fmt.Sprintf("%s %s %s", p.Tool, p.Action, strings.Join(p.Args, " "))
}

// gate returns the documented capability error for a validated plan.
func gate(p Plan) error {
	return command.Failuref(
		"%s: planned action [%s] requires privileged kernel network configuration not available in this "+
			"environment (capability-gated backend)", p.Tool, p.String())
}

// gatedNotes are the shared help notes for every capability-gated netctl applet.
var gatedNotes = []string{
	"Capability-gated: applying the plan needs CAP_NET_ADMIN and kernel support that MimixBox does not exercise, so the command reports a backend error instead of changing kernel state.",
}

// gatedNote is the shared help description suffix for every netctl applet.
const gatedNote = "This applet reconfigures privileged kernel network state, which is not available in this " +
	"environment. It validates arguments and serializes the requested action into a plan, then fails " +
	"with a documented capability error (it never silently does nothing)."

// gatedExitStatus is the shared ExitStatus help text for every netctl applet.
const gatedExitStatus = "0  never (capability-gated).\n1  always: validated plan then a documented backend error."

// Command is one of the netctl applets, selected by name.
type Command struct {
	name string
}

// NewBrctl returns a brctl command.
func NewBrctl() *Command { return &Command{name: "brctl"} }

// NewIfenslave returns an ifenslave command.
func NewIfenslave() *Command { return &Command{name: "ifenslave"} }

// NewTunctl returns a tunctl command.
func NewTunctl() *Command { return &Command{name: "tunctl"} }

// NewVconfig returns a vconfig command.
func NewVconfig() *Command { return &Command{name: "vconfig"} }

// NewZcip returns a zcip command.
func NewZcip() *Command { return &Command{name: "zcip"} }

// NewNbdClient returns an nbd-client command.
func NewNbdClient() *Command { return &Command{name: "nbd-client"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "brctl":
		return "Manage Ethernet bridges (capability-gated)"
	case "ifenslave":
		return "Attach/detach bonding slaves (capability-gated)"
	case "tunctl":
		return "Create/delete TUN/TAP devices (capability-gated)"
	case "vconfig":
		return "Manage 802.1q VLAN interfaces (capability-gated)"
	case "zcip":
		return "Manage IPv4 link-local addresses (capability-gated)"
	case "nbd-client":
		return "Attach a network block device (capability-gated)"
	}
	return "networking control"
}

// Run validates args, builds the plan, and returns the capability error via the
// shared plan-and-gate flow.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, c.usage(), stdio.Err).WithHelp(c.help())
	var built Plan
	return command.PlanGate{
		Plan: func(operands []string) (string, error) {
			p, err := c.plan(operands)
			if err != nil {
				return "", err
			}
			built = p
			return p.String(), nil
		},
		Gate: func(string) error { return gate(built) },
	}.Run(fs, stdio, args)
}

// plan validates operands and builds the Plan for the applet.
func (c *Command) plan(args []string) (Plan, error) {
	switch c.name {
	case "brctl":
		return planBrctl(args)
	case "ifenslave":
		return planIfenslave(args)
	case "tunctl":
		return planTunctl(args)
	case "vconfig":
		return planVconfig(args)
	case "zcip":
		return planZcip(args)
	case "nbd-client":
		return planNbdClient(args)
	}
	return Plan{}, fmt.Errorf("%s: unknown applet", c.name)
}

// usage returns the synopsis line for the applet's flag set.
func (c *Command) usage() string {
	switch c.name {
	case "brctl":
		return brctlUsage
	case "ifenslave":
		return ifenslaveUsage
	case "tunctl":
		return tunctlUsage
	case "vconfig":
		return vconfigUsage
	case "zcip":
		return zcipUsage
	case "nbd-client":
		return nbdClientUsage
	}
	return "[ARG...]"
}

// help returns the full help text for the applet.
func (c *Command) help() command.Help {
	switch c.name {
	case "brctl":
		return brctlHelp()
	case "ifenslave":
		return ifenslaveHelp()
	case "tunctl":
		return tunctlHelp()
	case "vconfig":
		return vconfigHelp()
	case "zcip":
		return zcipHelp()
	case "nbd-client":
		return nbdClientHelp()
	}
	return command.Help{Description: gatedNote}
}
