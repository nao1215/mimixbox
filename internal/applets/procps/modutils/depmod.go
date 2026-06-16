package modutils

import (
	"fmt"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// runDepmod is the CLI surface for the depmod applet: it builds the dependency
// plan with the shared planner, then either prints it (-n) or reports the gated
// install.
func (c *Command) runDepmod(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "[-n] [DIR]", stdio.Err).WithHelp(command.Help{
		Description: "Scan a directory of .ko modules and report the dependency relationships derived from " +
			"each module's .modinfo 'depends' field. With -n the dependency map is written to standard " +
			"output (a dry run) instead of installing modules.dep. The default behaviour, writing into " +
			"the system module directory, requires write access to /lib/modules and is intentionally " +
			"gated; pass a DIR and -n to run the hermetic analysis.",
		Examples: []command.Example{
			{Command: "depmod -n ./modules", Explain: "Print the dependency map for the .ko files under ./modules."},
		},
		ExitStatus: "0  with -n and a readable DIR.\n1  the directory is unreadable, or installation (without -n) is gated.",
		Notes:      []string{"Installing modules.dep into /lib/modules is gated; use -n for the analysis-only path."},
	})
	dryRun := fs.BoolP("dry-run", "n", false, "print the dependency map instead of installing it")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	dir := "."
	if rest := fs.Args(); len(rest) > 0 {
		dir = rest[0]
	}
	deps, err := buildDeps(dir)
	if err != nil {
		fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
		return command.SilentFailure()
	}
	if !*dryRun {
		return command.Failuref(
			"%s: dependency analysis succeeded for %d module(s), but installing modules.dep into "+
				"/lib/modules requires write access there; rerun with -n to print the map instead", c.name, len(deps))
	}
	for _, d := range deps {
		if len(d.deps) == 0 {
			fmt.Fprintf(stdio.Out, "%s:\n", d.path)
			continue
		}
		fmt.Fprintf(stdio.Out, "%s: %s\n", d.path, strings.Join(d.deps, " "))
	}
	return nil
}
