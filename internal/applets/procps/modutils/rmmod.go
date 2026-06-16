package modutils

import (
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// runRmmod is the CLI surface for the rmmod applet: it validates the module
// names with the shared planner, then reports the gated removal.
func (c *Command) runRmmod(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "MODULE...", stdio.Err).WithHelp(command.Help{
		Description: "Remove (unload) kernel modules by name. The module name is validated (a bare name " +
			"without slashes or a .ko suffix), then removal is attempted. Kernel removal requires " +
			"CAP_SYS_MODULE and is intentionally gated, failing with a documented error.",
		Examples:   []command.Example{{Command: "rmmod loop", Explain: "Validate the name and report the gated removal."}},
		ExitStatus: "1  always in this build (privileged removal is gated); 1 also on an invalid module name.",
		Notes:      []string{"Kernel removal requires CAP_SYS_MODULE; only name validation is performed here."},
	})
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
	return gatedError(c.name, strings.Join(names, ", "))
}
