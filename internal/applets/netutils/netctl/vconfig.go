package netctl

import (
	"fmt"
	"strconv"

	"github.com/nao1215/mimixbox/internal/command"
)

// vconfigUsage is the synopsis line for the vconfig applet.
const vconfigUsage = "COMMAND [ARG...]"

// planVconfig validates vconfig operands and builds the Plan.
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

// vconfigHelp returns the full help text for the vconfig applet.
func vconfigHelp() command.Help {
	return command.Help{
		Description: "Manage 802.1q VLAN interfaces (add, rem, set_flag, ...). " + gatedNote,
		Examples:    []command.Example{{Command: "vconfig add eth0 100", Explain: "Plan creating VLAN 100 on eth0."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
