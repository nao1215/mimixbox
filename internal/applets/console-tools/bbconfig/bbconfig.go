// Package bbconfig implements the bbconfig applet: print the MimixBox build
// configuration, that is the set of compiled-in applets, in a BusyBox-style
// CONFIG_* form.
package bbconfig

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/version"
)

// Command is the bbconfig applet.
type Command struct{}

// New returns a bbconfig command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "bbconfig" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Print the MimixBox build configuration" }

// appletNamesFn is indirected so the configuration source can be supplied in a
// test without re-invoking the binary. In production it asks the running
// MimixBox binary for its applet list.
var appletNamesFn = defaultAppletNames

// defaultAppletNames returns every compiled-in applet name by re-invoking the
// running MimixBox binary with --list (the same self-dispatch the busybox
// front-end uses) and reading the "name - description" table it prints.
func defaultAppletNames() ([]string, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("cannot locate the MimixBox binary: %w", err)
	}
	var out bytes.Buffer
	cmd := exec.Command(self, "--list") //nolint:gosec // re-invoking our own binary
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cannot read the applet list: %w", err)
	}
	return parseList(&out), nil
}

// parseList extracts the applet names from the "  name - description" table
// printed by `mimixbox --list`.
func parseList(r *bytes.Buffer) []string {
	var names []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		name, _, ok := strings.Cut(line, " - ")
		if !ok {
			name = strings.Fields(line)[0]
		}
		if name = strings.TrimSpace(name); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// Run executes bbconfig.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Print the MimixBox build configuration: the version and the list of compiled-in " +
			"applets in a BusyBox-style 'CONFIG_<APPLET>=y' form, one per line and sorted by name. " +
			"With -n / --names only the bare applet names are printed, one per line. This is the " +
			"MimixBox analogue of BusyBox's embedded .config; the list is read from the running " +
			"binary, so it always reflects what is actually compiled in.",
		Examples: []command.Example{
			{Command: "bbconfig", Explain: "Print CONFIG_* lines for every applet."},
			{Command: "bbconfig --names", Explain: "Print just the applet names, one per line."},
		},
		ExitStatus: "0  the configuration was printed.\n1  an unexpected argument was given or the list could not be read.",
	})
	namesOnly := fs.BoolP("names", "n", false, "print only the bare applet names, one per line")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 {
		return command.Failuref("unexpected argument: %q", rest[0])
	}

	names, err := appletNamesFn()
	if err != nil {
		return command.Failuref("%v", err)
	}
	sort.Strings(names)

	if *namesOnly {
		for _, n := range names {
			_, _ = fmt.Fprintln(stdio.Out, n)
		}
		return nil
	}

	_, _ = fmt.Fprintf(stdio.Out, "# MimixBox %s build configuration\n", version.Version)
	_, _ = fmt.Fprintf(stdio.Out, "CONFIG_MIMIXBOX_VERSION=%q\n", version.Version)
	_, _ = fmt.Fprintf(stdio.Out, "CONFIG_NUM_APPLETS=%d\n", len(names))
	for _, n := range names {
		_, _ = fmt.Fprintf(stdio.Out, "CONFIG_%s=y\n", configName(n))
	}
	return nil
}

// configName turns an applet name into a CONFIG_* symbol fragment: uppercased
// with any non-alphanumeric character replaced by an underscore, mirroring the
// Kconfig symbol naming BusyBox uses.
func configName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - ('a' - 'A'))
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
