package linkadmin

import "github.com/nao1215/mimixbox/internal/command"

// NewTC returns the tc applet. tc is inspect-only in this slice: show/list over
// an OBJECT (qdisc/class/filter) report no entries; mutating subcommands defer.
func NewTC() *Command {
	return &Command{
		spec: spec{
			name:     "tc",
			synopsis: "Show/parse traffic control configuration (inspect-only)",
			usage:    "OBJECT { show | list } [dev DEV]",
			desc: "Inspect Linux traffic control (qdisc/class/filter) configuration. This slice implements " +
				"the read-only show/list subcommands, which report no configured entries in the hermetic " +
				"environment; mutating subcommands (add/change/del) are intentionally deferred.",
		},
		examples: []command.Example{
			{Command: "tc qdisc show", Explain: "List qdiscs (empty in this slice)."},
		},
		run: runTC,
	}
}

// runTC dispatches the inspect/mutating decision for tc. The first operand is an
// OBJECT (qdisc/class/filter) and the subcommand follows it; a bare object
// (e.g. "tc qdisc") defaults to show.
func runTC(stdio command.IO, operands []string) error {
	sub := ""
	if len(operands) >= 2 {
		sub = operands[1]
	}
	switch sub {
	case "show", "list", "":
		return inspectNotice(stdio, "tc")
	default:
		return deferMutating(sub)
	}
}
