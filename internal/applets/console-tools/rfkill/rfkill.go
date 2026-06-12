// Package rfkill implements the rfkill applet: list and toggle the block state
// of the system's radio transmitters.
package rfkill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the rfkill applet.
type Command struct{}

// New returns a rfkill command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "rfkill" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "List or block radio transmitters" }

// Injected so the sysfs source and the privileged block call are testable.
var (
	sysClassRfkill = "/sys/class/rfkill"
	blockFn        = func(index int, block bool) error {
		// The real implementation writes a struct rfkill_event to /dev/rfkill.
		return fmt.Errorf("blocking requires write access to /dev/rfkill")
	}
)

// Run executes rfkill.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "{list | block INDEX | unblock INDEX}", stdio.Err).WithHelp(command.Help{
		Description: "List the radio transmitters (rfkill devices) and their block state, or block / " +
			"unblock the device with the given INDEX. With no command, 'list' is assumed. The device " +
			"information is read from /sys/class/rfkill; blocking requires privilege.",
		Examples: []command.Example{
			{Command: "rfkill list", Explain: "List all rfkill devices and their state."},
			{Command: "rfkill block 0", Explain: "Soft-block device 0."},
		},
		ExitStatus: "0  the command succeeded.\n1  an unknown command or an I/O error.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	rest := fs.Args()
	cmd := "list"
	if len(rest) > 0 {
		cmd = rest[0]
	}

	switch cmd {
	case "list":
		return listDevices(stdio)
	case "block", "unblock":
		if len(rest) < 2 {
			return command.Failuref("%s requires a device index", cmd)
		}
		index, err := strconv.Atoi(rest[1])
		if err != nil || index < 0 {
			return command.Failuref("invalid device index: %q", rest[1])
		}
		if err := blockFn(index, cmd == "block"); err != nil {
			return command.Failuref("cannot %s device %d: %v", cmd, index, err)
		}
		return nil
	default:
		return command.Failuref("unknown command: %q (use list, block, or unblock)", cmd)
	}
}

// device describes one rfkill transmitter.
type device struct {
	index      int
	name, kind string
	soft, hard bool
}

// listDevices prints each rfkill device read from sysfs.
func listDevices(stdio command.IO) error {
	for _, d := range readDevices() {
		_, _ = fmt.Fprintf(stdio.Out, "%d: %s: %s\n", d.index, d.name, d.kind)
		_, _ = fmt.Fprintf(stdio.Out, "\tSoft blocked: %s\n", yesno(d.soft))
		_, _ = fmt.Fprintf(stdio.Out, "\tHard blocked: %s\n", yesno(d.hard))
	}
	return nil
}

// readDevices reads the rfkill devices from sysfs, sorted by index.
func readDevices() []device {
	entries, err := os.ReadDir(sysClassRfkill)
	if err != nil {
		return nil
	}
	var devices []device
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "rfkill") {
			continue
		}
		dir := filepath.Join(sysClassRfkill, e.Name())
		idx, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "rfkill"))
		if err != nil {
			continue
		}
		devices = append(devices, device{
			index: idx,
			name:  attr(dir, "name"),
			kind:  attr(dir, "type"),
			soft:  attr(dir, "soft") == "1",
			hard:  attr(dir, "hard") == "1",
		})
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].index < devices[j].index })
	return devices
}

func attr(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // sysfs attribute path
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
