package netctl

import (
	"fmt"

	"github.com/nao1215/mimixbox/internal/command"
)

// ifenslaveUsage is the synopsis line for the ifenslave applet.
const ifenslaveUsage = "MASTER SLAVE..."

// planIfenslave validates ifenslave operands and builds the Plan.
func planIfenslave(args []string) (Plan, error) {
	if len(args) < 2 {
		return Plan{}, fmt.Errorf("ifenslave: a master and at least one slave interface are required")
	}
	return Plan{Tool: "ifenslave", Action: "enslave", Args: args}, nil
}

// ifenslaveHelp returns the full help text for the ifenslave applet.
func ifenslaveHelp() command.Help {
	return command.Help{
		Description: "Attach (or detach) slave interfaces to a bonding master. " + gatedNote,
		Examples:    []command.Example{{Command: "ifenslave bond0 eth0 eth1", Explain: "Plan enslaving eth0 and eth1 to bond0."}},
		Notes:       gatedNotes,
		ExitStatus:  gatedExitStatus,
	}
}
