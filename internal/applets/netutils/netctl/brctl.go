package netctl

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// brctlUsage is the synopsis line for the brctl applet.
const brctlUsage = "COMMAND [BRIDGE [INTERFACE]]"

// planBrctl validates brctl operands and builds the Plan.
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

// brctlHelp returns the full help text for the brctl applet.
func brctlHelp() command.Help {
	return command.Help{
		Description: "Manage Ethernet bridges (addbr, delbr, addif, delif, show). " + gatedNote,
		Examples:    []command.Example{{Command: "brctl addbr br0", Explain: "Plan creating bridge br0."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
