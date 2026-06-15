// Package lsmod implements the lsmod applet: list the kernel modules that are
// currently loaded, parsed from /proc/modules and formatted like the classic
// module-init-tools lsmod.
package lsmod

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the lsmod applet.
type Command struct{}

// New returns an lsmod command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "lsmod" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List loaded kernel modules" }

// modulesPath is the source for the module table; tests point it at a fixture.
var modulesPath = "/proc/modules"

// module is one row parsed from /proc/modules.
type module struct {
	name    string
	size    int64
	usedBy  int
	usedSet []string
}

// Run executes lsmod.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "List the kernel modules that are currently loaded, read from /proc/modules. " +
			"Each line shows the module name, its size in bytes, the number of other modules or " +
			"users referencing it, and (when present) the list of modules that depend on it. " +
			"This is a read-only query: it never inserts or removes modules.",
		Examples: []command.Example{
			{Command: "lsmod", Explain: "List all loaded modules."},
		},
		ExitStatus: "0  success.\n1  /proc/modules could not be read.",
		Notes: []string{
			"Reads /proc/modules; on systems without a Linux module subsystem the file is absent and lsmod fails with a clear error.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	f, err := os.Open(modulesPath) //nolint:gosec // fixed /proc path
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: cannot read %s: %v\n", c.Name(), modulesPath, err)
		return command.SilentFailure()
	}
	defer func() { _ = f.Close() }()

	mods, err := parseModules(f)
	if err != nil {
		_, _ = fmt.Fprintf(stdio.Err, "%s: %v\n", c.Name(), err)
		return command.SilentFailure()
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "Module\tSize\tUsed by")
	for _, m := range mods {
		usedBy := ""
		if len(m.usedSet) > 0 {
			usedBy = strings.Join(m.usedSet, ",")
		}
		_, _ = fmt.Fprintf(tw, "%s\t%d\t%d %s\n", m.name, m.size, m.usedBy, usedBy)
	}
	return tw.Flush()
}

// parseModules parses /proc/modules content. Each line looks like:
//
//	ext4 1024000 1 - Live 0xffffffffc0000000
//	mbcache 16384 1 ext4, Live 0xffffffffc0010000
func parseModules(r io.Reader) ([]module, error) {
	var mods []module
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("malformed /proc/modules line: %q", line)
		}
		var size int64
		if _, err := fmt.Sscanf(fields[1], "%d", &size); err != nil {
			return nil, fmt.Errorf("malformed module size in line: %q", line)
		}
		var used int
		if _, err := fmt.Sscanf(fields[2], "%d", &used); err != nil {
			return nil, fmt.Errorf("malformed use count in line: %q", line)
		}
		m := module{name: fields[0], size: size, usedBy: used}
		// fields[3] is the dependency list, "-" when empty, otherwise a
		// comma-separated list with a trailing comma.
		if len(fields) >= 4 && fields[3] != "-" {
			dep := strings.TrimSuffix(fields[3], ",")
			for _, d := range strings.Split(dep, ",") {
				if d != "" {
					m.usedSet = append(m.usedSet, d)
				}
			}
		}
		mods = append(mods, m)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return mods, nil
}
