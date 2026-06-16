package linkadmin

import "github.com/nao1215/mimixbox/internal/command"

// NewIPTunnel returns the iptunnel applet. It is inspect-only in this slice:
// show/list report no entries; creating/changing/deleting tunnels is deferred.
func NewIPTunnel() *Command {
	return &Command{
		spec: spec{
			name:     "iptunnel",
			synopsis: "Show/parse IP tunnels (inspect-only)",
			usage:    "{ show | list } [NAME]",
			desc: "Inspect IP tunnels (ipip/sit/gre). This slice implements the read-only show/list " +
				"subcommands; creating, changing, or deleting tunnels is intentionally deferred.",
		},
		examples: []command.Example{
			{Command: "iptunnel show", Explain: "List tunnels (empty in this slice)."},
		},
		run: runIPTunnel,
	}
}

// runIPTunnel dispatches the inspect/mutating decision for iptunnel. The first
// operand is the subcommand; absent, it defaults to show.
func runIPTunnel(stdio command.IO, operands []string) error {
	sub := ""
	if len(operands) > 0 {
		sub = operands[0]
	}
	switch sub {
	case "show", "list", "":
		return inspectNotice(stdio, "iptunnel")
	default:
		return deferMutating(sub)
	}
}
