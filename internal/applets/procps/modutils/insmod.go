package modutils

import (
	"fmt"
	"path/filepath"

	"github.com/nao1215/mimixbox/internal/command"
)

// runInsmod is the CLI surface for the insmod applet: it validates a module
// file with the shared planner, then reports the gated insertion.
func (c *Command) runInsmod(stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.name, "MODULE.ko [params]", stdio.Err).WithHelp(command.Help{
		Description: "Validate a kernel module file (MODULE.ko) by confirming it is an ELF object with a " +
			"loadable module section, then attempt to insert it. The validation/plan step is performed " +
			"locally; the actual kernel insertion requires CAP_SYS_MODULE and is intentionally gated, " +
			"failing with a documented error instead of silently doing nothing.",
		Examples:   []command.Example{{Command: "insmod ./loop.ko", Explain: "Validate loop.ko and report the gated insertion."}},
		ExitStatus: "1  always in this build (privileged insertion is gated); 1 also on a missing or invalid module file.",
		Notes:      []string{"Kernel insertion requires CAP_SYS_MODULE; only validation is performed here."},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintf(stdio.Err, "%s: no module file given\n", c.name)
		return command.SilentFailure()
	}
	mod := files[0]
	if err := validateModuleFile(mod); err != nil {
		fmt.Fprintf(stdio.Err, "%s: %s: %v\n", c.name, mod, err)
		return command.SilentFailure()
	}
	return gatedError(c.name, filepath.Base(mod))
}
