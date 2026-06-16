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
	"net"
	"strconv"
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

func planBrctl(args []string) (Plan, error) {
	if len(args) < 1 {
		return Plan{}, fmt.Errorf("brctl: a command is required (addbr/delbr/addif/delif/show)")
	}
	action := args[0]
	rest := args[1:]
	switch action {
	case "addbr", "delbr", "show":
		if action != "show" && len(rest) < 1 {
			return Plan{}, fmt.Errorf("brctl %s: a bridge name is required", action)
		}
	case "addif", "delif":
		if len(rest) < 2 {
			return Plan{}, fmt.Errorf("brctl %s: a bridge and an interface are required", action)
		}
	default:
		return Plan{}, fmt.Errorf("brctl: unknown command %q", action)
	}
	return Plan{Tool: "brctl", Action: action, Args: rest}, nil
}

func planIfenslave(args []string) (Plan, error) {
	if len(args) < 2 {
		return Plan{}, fmt.Errorf("ifenslave: a master and at least one slave interface are required")
	}
	return Plan{Tool: "ifenslave", Action: "enslave", Args: args}, nil
}

func planTunctl(args []string) (Plan, error) {
	// tunctl [-d IFACE] (delete) or tunctl [-t IFACE] (create).
	action := "create"
	var name string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			action = "delete"
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "-t":
			action = "create"
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		default:
			if name == "" {
				name = args[i]
			}
		}
	}
	if name == "" {
		return Plan{}, fmt.Errorf("tunctl: an interface name is required (-t NAME or -d NAME)")
	}
	return Plan{Tool: "tunctl", Action: action, Args: []string{name}}, nil
}

func planVconfig(args []string) (Plan, error) {
	if len(args) < 1 {
		return Plan{}, fmt.Errorf("vconfig: a command is required (add/rem/set_flag/...)")
	}
	action := args[0]
	rest := args[1:]
	switch action {
	case "add":
		if len(rest) < 2 {
			return Plan{}, fmt.Errorf("vconfig add: IFACE and VID are required")
		}
		vid, err := strconv.Atoi(rest[1])
		if err != nil || vid < 0 || vid > 4094 {
			return Plan{}, fmt.Errorf("vconfig add: invalid VLAN id %q (0-4094)", rest[1])
		}
	case "rem":
		if len(rest) < 1 {
			return Plan{}, fmt.Errorf("vconfig rem: a VLAN interface name is required")
		}
	default:
		if len(rest) < 1 {
			return Plan{}, fmt.Errorf("vconfig %s: an argument is required", action)
		}
	}
	return Plan{Tool: "vconfig", Action: action, Args: rest}, nil
}

func planZcip(args []string) (Plan, error) {
	if len(args) < 2 {
		return Plan{}, fmt.Errorf("zcip: an interface and a script are required")
	}
	return Plan{Tool: "zcip", Action: "configure", Args: args}, nil
}

func planNbdClient(args []string) (Plan, error) {
	// nbd-client HOST PORT NBDDEVICE  (or -d NBDDEVICE to disconnect).
	if len(args) >= 2 && args[0] == "-d" {
		return Plan{Tool: "nbd-client", Action: "disconnect", Args: args[1:2]}, nil
	}
	if len(args) < 3 {
		return Plan{}, fmt.Errorf("nbd-client: HOST PORT NBDDEVICE are required (or -d NBDDEVICE)")
	}
	if net.ParseIP(args[0]) == nil && !looksLikeHostname(args[0]) {
		return Plan{}, fmt.Errorf("nbd-client: invalid host %q", args[0])
	}
	if p, err := strconv.Atoi(args[1]); err != nil || p < 1 || p > 65535 {
		return Plan{}, fmt.Errorf("nbd-client: invalid port %q", args[1])
	}
	return Plan{Tool: "nbd-client", Action: "connect", Args: args[:3]}, nil
}

func looksLikeHostname(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t/") {
		return false
	}
	return true
}

func (c *Command) usage() string {
	switch c.name {
	case "brctl":
		return "COMMAND [BRIDGE [INTERFACE]]"
	case "ifenslave":
		return "MASTER SLAVE..."
	case "tunctl":
		return "[-t NAME | -d NAME]"
	case "vconfig":
		return "COMMAND [ARG...]"
	case "zcip":
		return "IFACE SCRIPT"
	case "nbd-client":
		return "HOST PORT NBDDEVICE | -d NBDDEVICE"
	}
	return "[ARG...]"
}

func (c *Command) help() command.Help {
	note := "This applet reconfigures privileged kernel network state, which is not available in this " +
		"environment. It validates arguments and serializes the requested action into a plan, then fails " +
		"with a documented capability error (it never silently does nothing)."
	gatedNotes := []string{
		"Capability-gated: applying the plan needs CAP_NET_ADMIN and kernel support that MimixBox does not exercise, so the command reports a backend error instead of changing kernel state.",
	}
	switch c.name {
	case "brctl":
		return command.Help{
			Description: "Manage Ethernet bridges (addbr, delbr, addif, delif, show). " + note,
			Examples:    []command.Example{{Command: "brctl addbr br0", Explain: "Plan creating bridge br0."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	case "ifenslave":
		return command.Help{
			Description: "Attach (or detach) slave interfaces to a bonding master. " + note,
			Examples:    []command.Example{{Command: "ifenslave bond0 eth0 eth1", Explain: "Plan enslaving eth0 and eth1 to bond0."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	case "tunctl":
		return command.Help{
			Description: "Create (-t) or delete (-d) a persistent TUN/TAP device. " + note,
			Examples:    []command.Example{{Command: "tunctl -t tap0", Explain: "Plan creating TAP device tap0."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	case "vconfig":
		return command.Help{
			Description: "Manage 802.1q VLAN interfaces (add, rem, set_flag, ...). " + note,
			Examples:    []command.Example{{Command: "vconfig add eth0 100", Explain: "Plan creating VLAN 100 on eth0."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	case "zcip":
		return command.Help{
			Description: "Manage IPv4 link-local (169.254/16) addressing via a configuration script. " + note,
			Examples:    []command.Example{{Command: "zcip eth0 /etc/zcip.script", Explain: "Plan link-local configuration of eth0."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	case "nbd-client":
		return command.Help{
			Description: "Attach (HOST PORT NBDDEVICE) or detach (-d NBDDEVICE) a network block device. " + note,
			Examples:    []command.Example{{Command: "nbd-client 10.0.0.1 10809 /dev/nbd0", Explain: "Plan attaching /dev/nbd0 to 10.0.0.1:10809."}},
			Notes:       gatedNotes,
			ExitStatus:  "0  never (capability-gated).\n1  always: validated plan then a documented backend error.",
		}
	}
	return command.Help{Description: note}
}
