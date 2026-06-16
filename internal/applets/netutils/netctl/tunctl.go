package netctl

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// tunctlUsage is the synopsis line for the tunctl applet.
const tunctlUsage = "[-t NAME | -d NAME]"

// planTunctl validates tunctl operands and builds the Plan.
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

// tunctlHelp returns the full help text for the tunctl applet.
func tunctlHelp() command.Help {
	return command.Help{
		Description: "Create (-t) or delete (-d) a persistent TUN/TAP device. " + gatedNote,
		Examples:    []command.Example{{Command: "tunctl -t tap0", Explain: "Plan creating TAP device tap0."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
