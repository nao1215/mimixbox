package linkadmin

import "github.com/nao1215/mimixbox/internal/command"

// NewSlattach returns the slattach applet. Attaching a serial line needs
// privileged device access and is deferred: slattach accepts its -p protocol
// option and validates its TTY arity, then reports a capability error.
func NewSlattach() *Command {
	return &Command{
		spec: spec{
			name:     "slattach",
			synopsis: "Attach a serial line as a network interface (deferred)",
			usage:    "[-p PROTOCOL] TTY",
			desc: "Attach a serial line (TTY) as a SLIP/PPP network interface. This needs privileged access " +
				"to the serial device and is intentionally deferred; the command validates its arguments " +
				"and then reports a capability error.",
		},
		examples: []command.Example{
			{Command: "slattach -p slip /dev/ttyS0", Explain: "Attach SLIP (deferred)."},
		},
		// slattach takes a -p protocol option; accept it so parsing succeeds.
		addFlags: func(fs *command.FlagSet) {
			_ = fs.StringP("protocol", "p", "slip", "line protocol (validated, then deferred)")
		},
		run: runSlattach,
	}
}

func runSlattach(_ command.IO, operands []string) error {
	if len(operands) != 1 {
		return command.Failuref("usage: slattach [-p PROTOCOL] TTY")
	}
	return command.Failuref(
		"attaching serial line %q requires privileged device access and is intentionally "+
			"deferred in this slice", operands[0])
}
