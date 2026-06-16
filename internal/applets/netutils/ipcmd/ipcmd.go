// Package ipcmd implements the iproute2-style ip family of applets (ip, ipaddr,
// iplink, iproute, ipneigh, iprule). They all share one parser and one
// read-only backend: each command resolves to an object (link/address/route/
// neighbour/rule) and a subcommand (show/list), then renders fixture data from
// an injectable source. Mutating subcommands (add/del/...) are intentionally
// deferred and fail with a documented capability error, never a silent no-op.
//
// This file is the command surface: the per-applet constructors, names,
// synopses, usage strings, and help text. The shared parsing/dump backend that
// every command funnels through lives in backend.go.
package ipcmd

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is one applet in the ip family.
type Command struct {
	name string
	// fixedObject, when non-nil, pins the object so dedicated applets such as
	// ipaddr behave like "ip addr" without taking an object argument.
	fixedObject *object
}

// NewIP returns the generic "ip" applet which takes OBJECT as its first operand.
func NewIP() *Command { return &Command{name: "ip"} }

// NewIPAddr returns "ipaddr" (equivalent to "ip address").
func NewIPAddr() *Command { o := objAddr; return &Command{name: "ipaddr", fixedObject: &o} }

// NewIPLink returns "iplink" (equivalent to "ip link").
func NewIPLink() *Command { o := objLink; return &Command{name: "iplink", fixedObject: &o} }

// NewIPRoute returns "iproute" (equivalent to "ip route").
func NewIPRoute() *Command { o := objRoute; return &Command{name: "iproute", fixedObject: &o} }

// NewIPNeigh returns "ipneigh" (equivalent to "ip neighbour").
func NewIPNeigh() *Command { o := objNeigh; return &Command{name: "ipneigh", fixedObject: &o} }

// NewIPRule returns "iprule" (equivalent to "ip rule").
func NewIPRule() *Command { o := objRule; return &Command{name: "iprule", fixedObject: &o} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch {
	case c.fixedObject == nil:
		return "Show and manage routing, devices, and tunnels (read-only show/list)"
	case *c.fixedObject == objAddr:
		return "Show protocol (IP) addresses on devices"
	case *c.fixedObject == objLink:
		return "Show network device link state"
	case *c.fixedObject == objRoute:
		return "Show the routing table"
	case *c.fixedObject == objNeigh:
		return "Show the ARP/neighbour table"
	default:
		return "Show routing policy rules"
	}
}

// Run executes an ip-family command.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), c.usage(), stdio.Err).WithHelp(c.help())
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	obj, rest, err := c.resolveObject(operands)
	if err != nil {
		return command.Failuref("%v", err)
	}

	sub, target := splitSub(rest)
	switch sub {
	case "show", "list", "lst", "":
		dump(stdio.Out, obj, target)
		return nil
	default:
		return command.Failuref(
			"%q is a mutating subcommand and is not implemented in this read-only slice; "+
				"only show/list are available", sub)
	}
}

// resolveObject determines which object the command targets. For dedicated
// applets the object is fixed; for the generic "ip" applet it is parsed (by
// prefix) from the first operand.
func (c *Command) resolveObject(operands []string) (object, []string, error) {
	if c.fixedObject != nil {
		return *c.fixedObject, operands, nil
	}
	if len(operands) == 0 {
		return 0, nil, fmt.Errorf("an OBJECT is required (one of: link, address, route, neighbour, rule)")
	}
	obj, ok := matchObject(operands[0])
	if !ok {
		return 0, nil, fmt.Errorf("unknown object %q (expected link, address, route, neighbour, or rule)", operands[0])
	}
	return obj, operands[1:], nil
}

func (c *Command) usage() string {
	if c.fixedObject == nil {
		return "OBJECT { show | list }"
	}
	return "{ show | list }"
}

func (c *Command) help() command.Help {
	return command.Help{
		Description: "iproute2-style network inspection. This slice implements the read-only " +
			"show/list subcommands over an injectable data source so it is hermetic and never " +
			"touches the live kernel tables. For the generic 'ip' applet the first operand selects " +
			"the OBJECT (link, address, route, neighbour, rule); the dedicated ipaddr/iplink/" +
			"iproute/ipneigh/iprule applets pin that object. Mutating subcommands (add, del, change, " +
			"flush, ...) are intentionally deferred and report a clear error rather than acting.",
		Examples: []command.Example{
			{Command: "ip addr show", Explain: "List addresses on every device."},
			{Command: "ip link show dev eth0", Explain: "Show one device's link state."},
			{Command: "iproute show", Explain: "Print the routing table."},
			{Command: "ipneigh show", Explain: "Print the neighbour (ARP) table."},
		},
		ExitStatus: "0  the requested table was printed.\n" +
			"1  bad object/subcommand, or a mutating subcommand was requested.",
		Notes: []string{
			"Mutating subcommands are not yet implemented; they fail deterministically.",
		},
	}
}
