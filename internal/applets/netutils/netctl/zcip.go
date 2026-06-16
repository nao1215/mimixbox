package netctl

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// zcipUsage is the synopsis line for the zcip applet.
const zcipUsage = "IFACE SCRIPT"

// planZcip validates zcip operands and builds the Plan.
func planZcip(args []string) (Plan, error) {
	if len(args) < 2 {
		return Plan{}, fmt.Errorf("zcip: an interface and a script are required")
	}
	return Plan{Tool: "zcip", Action: "configure", Args: args}, nil
}

// zcipHelp returns the full help text for the zcip applet.
func zcipHelp() command.Help {
	return command.Help{
		Description: "Manage IPv4 link-local (169.254/16) addressing via a configuration script. " + gatedNote,
		Examples:    []command.Example{{Command: "zcip eth0 /etc/zcip.script", Explain: "Plan link-local configuration of eth0."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
