package linkadmin

import "github.com/nao1215/mimixbox/internal/command"

// NewNameif returns the nameif applet. Renaming touches the live kernel and is
// deferred: nameif validates its INTERFACE/MACADDRESS arity, then reports a
// capability error rather than renaming anything.
func NewNameif() *Command {
	return &Command{
		spec: spec{
			name:     "nameif",
			synopsis: "Rename network interfaces by MAC (deferred)",
			usage:    "INTERFACE MACADDRESS",
			desc: "Rename a network interface to match a MAC address. Renaming touches the live kernel and " +
				"is intentionally deferred in this slice; the command validates its arguments and then " +
				"reports a capability error rather than renaming anything.",
		},
		examples: []command.Example{
			{Command: "nameif eth0 00:11:22:33:44:55", Explain: "Rename (deferred)."},
		},
		run: runNameif,
	}
}

func runNameif(_ command.IO, operands []string) error {
	if len(operands) != 2 {
		return command.Failuref("usage: nameif INTERFACE MACADDRESS")
	}
	return command.Failuref(
		"renaming interface %q to match %q requires privileged kernel access and is "+
			"intentionally deferred in this slice", operands[0], operands[1])
}
