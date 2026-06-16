package modutils

import (
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// runModprobe is the CLI surface for the modprobe applet: it validates the
// module names with the shared planner, then reports the gated load/unload.
func (c *Command) runModprobe(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-r] MODULE...", stdio.Err).WithHelp(command.Help{
		Description: "Resolve and load (or with -r, unload) a kernel module by name, including its " +
			"dependencies. The module name is validated and the dependency plan is reported; the actual " +
			"kernel load/unload requires CAP_SYS_MODULE and is intentionally gated with a documented " +
			"error. Dependency resolution against /lib/modules is not performed in this hermetic build.",
		Examples:   []command.Example{{Command: "modprobe loop", Explain: "Validate and report the gated load of loop."}},
		ExitStatus: "1  always in this build (privileged load/unload is gated); 1 also on an invalid module name.",
		Notes:      []string{"Kernel load/unload requires CAP_SYS_MODULE; only name validation is performed here."},
	})
	remove := fs.BoolP("remove", "r", false, "unload instead of load")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	names := fs.Args()
	if len(names) == 0 {
		fmt.Fprintf(stdio.Err, "%s: no module name given\n", c.name)
		return command.SilentFailure()
	}
	if err := validateModuleNames(names); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	action := "load"
	if *remove {
		action = "unload"
	}
	return gatedError(c.name, action+" of "+strings.Join(names, ", "))
}
