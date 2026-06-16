// Package powertop implements the powertop applet: a one-shot report of the
// system's power supplies read from /sys/class/power_supply.
package powertop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the powertop applet.
type Command struct{}

// New returns a powertop command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "powertop" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Report the system power supplies" }

// powerSupplyDir is the sysfs power-supply class; tests point it at a fixture.
var powerSupplyDir = "/sys/class/power_supply"

// Run executes powertop in one-shot report mode.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print a one-shot report of the system's power supplies (AC adapters and " +
			"batteries) read from /sys/class/power_supply: the AC online state and each battery's " +
			"charge and status.",
		Examples: []command.Example{
			{Command: "powertop", Explain: "Report the power supplies."},
		},
		Notes: []string{
			"The interactive power analysis and tunables of real powertop are not implemented.",
		},
		ExitStatus: "0  success.\n1  an error occurred (e.g. /sys/class/power_supply could not be read or an argument was invalid).",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	supplies := readSupplies()
	if len(supplies) == 0 {
		_, _ = fmt.Fprintln(stdio.Out, "powertop: no power supplies found")
		return nil
	}

	_, _ = fmt.Fprintln(stdio.Out, "Power supplies:")
	for _, s := range supplies {
		_, _ = fmt.Fprintf(stdio.Out, "  %s\n", s)
	}
	return nil
}

// readSupplies returns one formatted line per power-supply device, sorted by
// name.
func readSupplies() []string {
	entries, err := os.ReadDir(powerSupplyDir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	var lines []string
	for _, name := range names {
		dir := filepath.Join(powerSupplyDir, name)
		switch readAttr(dir, "type") {
		case "Battery":
			cap := readAttr(dir, "capacity")
			status := readAttr(dir, "status")
			lines = append(lines, fmt.Sprintf("%s (Battery): %s%% (%s)", name, orUnknown(cap), orUnknown(status)))
		case "Mains":
			state := "offline"
			if readAttr(dir, "online") == "1" {
				state = "online"
			}
			lines = append(lines, fmt.Sprintf("%s (AC): %s", name, state))
		case "":
			// Not a recognizable power supply; skip it.
		default:
			lines = append(lines, fmt.Sprintf("%s (%s)", name, readAttr(dir, "type")))
		}
	}
	return lines
}

// readAttr reads a single sysfs attribute file, trimming whitespace.
func readAttr(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // sysfs attribute path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func orUnknown(s string) string {
	if s == "" {
		return "?"
	}
	return s
}
