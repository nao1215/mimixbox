// Package linkadmin implements the link/tunnel administration applets whose
// write paths are not yet safe for CI: tc, iptunnel, nameif, and slattach. They
// share one shape: parse and validate arguments, then report a deterministic
// capability error explaining that the privileged operation is intentionally
// deferred. This honours the "never a silent no-op" rule. Read-only inspect
// subcommands ("show"/"list") are accepted and produce an explicit "no entries"
// notice rather than failing, since inspection is the safe first step.
package linkadmin

import (
	"context"
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// kind identifies which administration applet a Command is.
type kind int

const (
	kindTC kind = iota
	kindIPTunnel
	kindNameif
	kindSlattach
)

type spec struct {
	name       string
	synopsis   string
	usage      string
	desc       string
	hasInspect bool // whether show/list is a meaningful inspect subcommand
}

var specs = map[kind]spec{
	kindTC: {
		name: "tc", synopsis: "Show/parse traffic control configuration (inspect-only)",
		usage: "OBJECT { show | list } [dev DEV]", hasInspect: true,
		desc: "Inspect Linux traffic control (qdisc/class/filter) configuration. This slice implements " +
			"the read-only show/list subcommands, which report no configured entries in the hermetic " +
			"environment; mutating subcommands (add/change/del) are intentionally deferred.",
	},
	kindIPTunnel: {
		name: "iptunnel", synopsis: "Show/parse IP tunnels (inspect-only)",
		usage: "{ show | list } [NAME]", hasInspect: true,
		desc: "Inspect IP tunnels (ipip/sit/gre). This slice implements the read-only show/list " +
			"subcommands; creating, changing, or deleting tunnels is intentionally deferred.",
	},
	kindNameif: {
		name: "nameif", synopsis: "Rename network interfaces by MAC (deferred)",
		usage: "INTERFACE MACADDRESS",
		desc: "Rename a network interface to match a MAC address. Renaming touches the live kernel and " +
			"is intentionally deferred in this slice; the command validates its arguments and then " +
			"reports a capability error rather than renaming anything.",
	},
	kindSlattach: {
		name: "slattach", synopsis: "Attach a serial line as a network interface (deferred)",
		usage: "[-p PROTOCOL] TTY",
		desc: "Attach a serial line (TTY) as a SLIP/PPP network interface. This needs privileged access " +
			"to the serial device and is intentionally deferred; the command validates its arguments " +
			"and then reports a capability error.",
	},
}

// Command is one link/tunnel administration applet.
type Command struct{ kind kind }

// NewTC returns the tc applet.
func NewTC() *Command { return &Command{kind: kindTC} }

// NewIPTunnel returns the iptunnel applet.
func NewIPTunnel() *Command { return &Command{kind: kindIPTunnel} }

// NewNameif returns the nameif applet.
func NewNameif() *Command { return &Command{kind: kindNameif} }

// NewSlattach returns the slattach applet.
func NewSlattach() *Command { return &Command{kind: kindSlattach} }

// Name returns the command name.
func (c *Command) Name() string { return specs[c.kind].name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return specs[c.kind].synopsis }

// Run executes the applet.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	s := specs[c.kind]
	fs := command.NewFlagSet(c.Name(), s.usage, stdio.Err).WithHelp(command.Help{
		Description: s.desc,
		Examples:    c.examples(),
		ExitStatus: "0  an inspect subcommand printed its (possibly empty) result.\n" +
			"1  a mutating/privileged operation was requested, or arguments were invalid.",
		Notes: []string{"Mutating/privileged operations are intentionally deferred and fail deterministically."},
	})
	// slattach takes a -p protocol option; accept it so parsing succeeds.
	if c.kind == kindSlattach {
		_ = fs.StringP("protocol", "p", "slip", "line protocol (validated, then deferred)")
	}
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	switch c.kind {
	case kindTC, kindIPTunnel:
		return c.runInspect(stdio, operands)
	case kindNameif:
		if len(operands) != 2 {
			return command.Failuref("usage: nameif INTERFACE MACADDRESS")
		}
		return command.Failuref(
			"renaming interface %q to match %q requires privileged kernel access and is "+
				"intentionally deferred in this slice", operands[0], operands[1])
	default: // slattach
		if len(operands) != 1 {
			return command.Failuref("usage: slattach [-p PROTOCOL] TTY")
		}
		return command.Failuref(
			"attaching serial line %q requires privileged device access and is intentionally "+
				"deferred in this slice", operands[0])
	}
}

// runInspect handles the read-only show/list subcommands for tc and iptunnel.
func (c *Command) runInspect(stdio command.IO, operands []string) error {
	sub := inspectSub(c.kind, operands)
	switch sub {
	case "show", "list", "":
		fmt.Fprintf(stdio.Out, "%s: no entries (inspect-only slice; the live kernel is not queried)\n", c.Name())
		return nil
	default:
		return command.Failuref(
			"%q is a mutating subcommand and is intentionally deferred; only show/list are available", sub)
	}
}

// inspectSub locates the subcommand keyword for an inspect-style command. For tc
// the first operand is an OBJECT (qdisc/class/filter) and the subcommand follows;
// for iptunnel the first operand is the subcommand.
func inspectSub(k kind, operands []string) string {
	if len(operands) == 0 {
		return ""
	}
	if k == kindTC && len(operands) >= 2 {
		return operands[1]
	}
	if k == kindTC {
		// Only an object given (e.g. "tc qdisc"): default to show.
		return ""
	}
	return operands[0]
}

func (c *Command) examples() []command.Example {
	switch c.kind {
	case kindTC:
		return []command.Example{{Command: "tc qdisc show", Explain: "List qdiscs (empty in this slice)."}}
	case kindIPTunnel:
		return []command.Example{{Command: "iptunnel show", Explain: "List tunnels (empty in this slice)."}}
	case kindNameif:
		return []command.Example{{Command: "nameif eth0 00:11:22:33:44:55", Explain: "Rename (deferred)."}}
	default:
		return []command.Example{{Command: "slattach -p slip /dev/ttyS0", Explain: "Attach SLIP (deferred)."}}
	}
}
