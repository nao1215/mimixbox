package netctl

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// nbdClientUsage is the synopsis line for the nbd-client applet.
const nbdClientUsage = "HOST PORT NBDDEVICE | -d NBDDEVICE"

// planNbdClient validates nbd-client operands and builds the Plan.
func planNbdClient(args []string) (Plan, error) {
	// nbd-client HOST PORT NBDDEVICE  (or -d NBDDEVICE to disconnect).
	if len(args) >= 2 && args[0] == "-d" {
		return Plan{Tool: "nbd-client", Action: "disconnect", Args: args[1:2]}, nil
	}
	if len(args) < 3 {
		return Plan{}, fmt.Errorf("nbd-client: HOST PORT NBDDEVICE are required (or -d NBDDEVICE)")
	}
	if net.ParseIP(args[0]) == nil && !looksLikeHostname(args[0]) {
		return Plan{}, fmt.Errorf("nbd-client: invalid host %q", args[0])
	}
	if p, err := strconv.Atoi(args[1]); err != nil || p < 1 || p > 65535 {
		return Plan{}, fmt.Errorf("nbd-client: invalid port %q", args[1])
	}
	return Plan{Tool: "nbd-client", Action: "connect", Args: args[:3]}, nil
}

// looksLikeHostname reports whether s is a plausible hostname operand.
func looksLikeHostname(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t/") {
		return false
	}
	return true
}

// nbdClientHelp returns the full help text for the nbd-client applet.
func nbdClientHelp() command.Help {
	return command.Help{
		Description: "Attach (HOST PORT NBDDEVICE) or detach (-d NBDDEVICE) a network block device. " + gatedNote,
		Examples:    []command.Example{{Command: "nbd-client 10.0.0.1 10809 /dev/nbd0", Explain: "Plan attaching /dev/nbd0 to 10.0.0.1:10809."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
