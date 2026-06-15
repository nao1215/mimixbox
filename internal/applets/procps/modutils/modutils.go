// Package modutils implements the kernel-module mutation applets insmod, rmmod,
// modprobe, and depmod. Each separates metadata parsing / plan generation (which
// is hermetic and testable) from the privileged kernel mutation (which requires
// CAP_SYS_MODULE). The privileged step is intentionally gated: it fails
// deterministically with a documented requirement rather than silently doing
// nothing.
package modutils

import (
	"context"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

const (
	cmdInsmod   = "insmod"
	cmdRmmod    = "rmmod"
	cmdModprobe = "modprobe"
	cmdDepmod   = "depmod"
)

// Command is one module-mutation applet, distinguished by name.
type Command struct {
	name string
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// NewInsmod returns the insmod applet.
func NewInsmod() *Command { return &Command{name: cmdInsmod} }

// NewRmmod returns the rmmod applet.
func NewRmmod() *Command { return &Command{name: cmdRmmod} }

// NewModprobe returns the modprobe applet.
func NewModprobe() *Command { return &Command{name: cmdModprobe} }

// NewDepmod returns the depmod applet.
func NewDepmod() *Command { return &Command{name: cmdDepmod} }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case cmdInsmod:
		return "Validate and (privileged) insert a kernel module"
	case cmdRmmod:
		return "Validate and (privileged) remove a kernel module"
	case cmdModprobe:
		return "Resolve dependencies and (privileged) load a module"
	case cmdDepmod:
		return "Build the module dependency list"
	}
	return "Kernel module utility"
}

// Run dispatches to the per-applet implementation.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	switch c.name {
	case cmdInsmod:
		return c.runInsmod(stdio, args)
	case cmdRmmod:
		return c.runRmmod(stdio, args)
	case cmdModprobe:
		return c.runModprobe(stdio, args)
	case cmdDepmod:
		return c.runDepmod(stdio, args)
	}
	return command.Failuref("%s: unknown module applet", c.name)
}

// gatedError is the standard documented refusal for a privileged kernel
// mutation. metaOK indicates the metadata/plan step succeeded first.
func gatedError(name, detail string) error {
	return command.Failuref(
		"%s: %s validated successfully, but inserting/removing kernel modules requires CAP_SYS_MODULE; "+
			"this privileged step is intentionally not implemented in the hermetic build", name, detail)
}

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
	for _, n := range names {
		if err := validateModuleName(n); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			return command.SilentFailure()
		}
	}
	return gatedError(c.name, strings.Join(names, ", "))
}

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
	for _, n := range names {
		if err := validateModuleName(n); err != nil {
			fmt.Fprintf(stdio.Err, "%s: %v\n", c.name, err)
			return command.SilentFailure()
		}
	}
	action := "load"
	if *remove {
		action = "unload"
	}
	return gatedError(c.name, action+" of "+strings.Join(names, ", "))
}

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

// dep is one module and the modules it depends on, as parsed from .modinfo.
type dep struct {
	path string
	deps []string
}

// buildDeps scans dir for *.ko files and extracts each module's "depends"
// field, producing a dependency map. It is fully hermetic.
func buildDeps(dir string) ([]dep, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var deps []dep
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ko") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		field, err := modinfoField(full, "depends")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		var names []string
		if field != "" {
			for _, n := range strings.Split(field, ",") {
				if n != "" {
					names = append(names, n)
				}
			}
		}
		deps = append(deps, dep{path: e.Name(), deps: names})
	}
	return deps, nil
}

// validateModuleFile checks that path is an ELF object carrying a .modinfo
// section, the minimum that identifies a kernel module.
func validateModuleFile(path string) error {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	ef, err := elf.NewFile(f)
	if err != nil {
		return fmt.Errorf("not an ELF object: %w", err)
	}
	if ef.Section(".modinfo") == nil {
		return fmt.Errorf("missing .modinfo section (not a kernel module?)")
	}
	return nil
}

// validateModuleName checks that name is a bare module name, not a path.
func validateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("empty module name")
	}
	if strings.ContainsAny(name, "/") || strings.HasSuffix(name, ".ko") {
		return fmt.Errorf("%q is not a bare module name (use the module name, e.g. 'loop')", name)
	}
	return nil
}

// modinfoField returns one field value from a .ko file's .modinfo section.
func modinfoField(path, key string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // user-named file
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	ef, err := elf.NewFile(f)
	if err != nil {
		return "", fmt.Errorf("not an ELF object: %w", err)
	}
	sec := ef.Section(".modinfo")
	if sec == nil {
		return "", fmt.Errorf("no .modinfo section")
	}
	data, err := sec.Data()
	if err != nil {
		return "", err
	}
	for _, rec := range strings.Split(string(data), "\x00") {
		if i := strings.IndexByte(rec, '='); i >= 0 && rec[:i] == key {
			return rec[i+1:], nil
		}
	}
	return "", nil
}
