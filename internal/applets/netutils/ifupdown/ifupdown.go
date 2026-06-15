// Package ifupdown implements the ifup, ifdown, and ifplugd applets.
//
// The /etc/network/interfaces-style config parser is pure and table-tested. The
// applets parse a config file (a temp fixture in tests) and run the pre-up /
// up / down / post-down command hooks defined for an interface. The hook runner
// is injectable so tests can capture the planned commands instead of touching
// the host. Actually bringing a kernel interface up or down (the ip/ifconfig
// state change) is capability-gated and reported with a documented error.
package ifupdown

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Iface is one parsed "iface" stanza from an interfaces file.
type Iface struct {
	Name    string
	Family  string // inet, inet6, ...
	Method  string // dhcp, static, manual, loopback, ...
	Options map[string]string
	PreUp   []string
	Up      []string
	Down    []string
	PostDown []string
}

// Config is a parsed interfaces file: a set of iface stanzas plus auto names.
type Config struct {
	Auto   []string
	Ifaces map[string]*Iface
}

// ParseConfig parses an /etc/network/interfaces-style file from r.
func ParseConfig(r io.Reader) (*Config, error) {
	cfg := &Config{Ifaces: map[string]*Iface{}}
	sc := bufio.NewScanner(r)
	line := 0
	var cur *Iface
	for sc.Scan() {
		line++
		raw := sc.Text()
		text := strings.TrimSpace(raw)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		f := strings.Fields(text)
		switch f[0] {
		case "auto", "allow-hotplug":
			cfg.Auto = append(cfg.Auto, f[1:]...)
			cur = nil
		case "iface":
			if len(f) < 4 {
				return nil, fmt.Errorf("line %d: iface needs NAME FAMILY METHOD", line)
			}
			cur = &Iface{Name: f[1], Family: f[2], Method: f[3], Options: map[string]string{}}
			cfg.Ifaces[f[1]] = cur
		default:
			if cur == nil {
				return nil, fmt.Errorf("line %d: option %q outside an iface stanza", line, f[0])
			}
			val := strings.TrimSpace(strings.TrimPrefix(text, f[0]))
			switch f[0] {
			case "pre-up":
				cur.PreUp = append(cur.PreUp, val)
			case "up", "post-up":
				cur.Up = append(cur.Up, val)
			case "down", "pre-down":
				cur.Down = append(cur.Down, val)
			case "post-down":
				cur.PostDown = append(cur.PostDown, val)
			default:
				cur.Options[f[0]] = val
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// HookRunner runs a configured shell hook command. Injecting it keeps tests
// hermetic; the production runner executes the command with /bin/sh -c.
type HookRunner func(ctx context.Context, cmdline string) error

// Plan returns the ordered hook commands ifup (bringUp=true) or ifdown
// (bringUp=false) would run for the named interface.
func Plan(cfg *Config, name string, bringUp bool) ([]string, error) {
	ifc, ok := cfg.Ifaces[name]
	if !ok {
		return nil, fmt.Errorf("interface %q is not configured", name)
	}
	if bringUp {
		return append(append([]string{}, ifc.PreUp...), ifc.Up...), nil
	}
	return append(append([]string{}, ifc.Down...), ifc.PostDown...), nil
}

// RunHooks runs each command from Plan through runner, stopping at the first
// error.
func RunHooks(ctx context.Context, cmds []string, runner HookRunner) error {
	for _, c := range cmds {
		if strings.TrimSpace(c) == "" {
			continue
		}
		if err := runner(ctx, c); err != nil {
			return fmt.Errorf("hook %q failed: %w", c, err)
		}
	}
	return nil
}

// Command is the ifup / ifdown / ifplugd applet.
type Command struct {
	name string
}

// NewIfup returns an ifup command.
func NewIfup() *Command { return &Command{name: "ifup"} }

// NewIfdown returns an ifdown command.
func NewIfdown() *Command { return &Command{name: "ifdown"} }

// NewIfplugd returns an ifplugd command.
func NewIfplugd() *Command { return &Command{name: "ifplugd"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	switch c.name {
	case "ifdown":
		return "Take a network interface down"
	case "ifplugd":
		return "Bring interfaces up/down on link change"
	default:
		return "Bring a network interface up"
	}
}

// Run executes ifup/ifdown/ifplugd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if c.name == "ifplugd" {
		return runIfplugd(stdio, args)
	}
	return c.runIfupdown(ctx, stdio, args)
}

func (c *Command) runIfupdown(ctx context.Context, stdio command.IO, args []string) error {
	bringUp := c.name == "ifup"
	fs := command.NewFlagSet(c.name, "[-i FILE] [-n] IFACE", stdio.Err).WithHelp(command.Help{
		Description: "Parse an /etc/network/interfaces-style file and run the configured hook commands for " +
			"an interface. -i selects the interfaces file (default /etc/network/interfaces); use a temp " +
			"fixture for testing. -n (no-act) prints the hook commands that would run without running them " +
			"and without changing the interface. Bringing the kernel interface " + verb(bringUp) + " " +
			"requires privileged network configuration that is not available here, so without -n the " +
			"command runs the hooks and then reports a documented capability error for the state change.",
		Examples: []command.Example{
			{Command: c.name + " -n -i interfaces eth0", Explain: "Print the hook commands for eth0 without changing anything."},
		},
		ExitStatus: "0  with -n, or when hooks succeed and no state change is requested.\n1  config/hook error or capability-gated state change.",
		Notes: []string{
			"The kernel interface state change (ip link set) is capability-gated; hook execution and config parsing are implemented.",
		},
	})
	file := fs.StringP("interfaces", "i", "/etc/network/interfaces", "interfaces config file")
	noAct := fs.BoolP("no-act", "n", false, "print the hook commands without running them")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return command.Failuref("an interface name is required")
	}
	name := rest[0]

	f, err := os.Open(*file)
	if err != nil {
		return command.Failuref("cannot open %q: %v", *file, err)
	}
	defer func() { _ = f.Close() }()
	cfg, err := ParseConfig(f)
	if err != nil {
		return command.Failuref("%s: %v", *file, err)
	}

	cmds, err := Plan(cfg, name, bringUp)
	if err != nil {
		return command.Failure(err)
	}

	if *noAct {
		for _, c := range cmds {
			if strings.TrimSpace(c) != "" {
				_, _ = fmt.Fprintln(stdio.Out, c)
			}
		}
		return nil
	}

	if err := RunHooks(ctx, cmds, shellHook(stdio)); err != nil {
		return command.Failure(err)
	}
	return command.Failuref(
		"%s: ran hooks for %q, but bringing the interface %s requires privileged network configuration "+
			"not available in this environment (capability-gated backend)", c.name, name, verb(bringUp))
}

func runIfplugd(stdio command.IO, args []string) error {
	fs := command.NewFlagSet("ifplugd", "-i IFACE", stdio.Err).WithHelp(command.Help{
		Description: "Monitor an interface's link state and run ifup/ifdown when the cable is plugged or " +
			"unplugged. Link-state monitoring relies on privileged netlink/ioctl access that is not " +
			"available in this environment; this slice validates its arguments and reports a documented " +
			"capability error instead of silently doing nothing.",
		Examples: []command.Example{
			{Command: "ifplugd -i eth0", Explain: "Plan link monitoring of eth0 (capability-gated)."},
		},
		ExitStatus: "0  never in this environment.\n1  always: validated request then a documented backend error.",
	})
	iface := fs.StringP("iface", "i", "", "interface to monitor")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *iface == "" {
		return command.Failuref("an interface is required (-i)")
	}
	return command.Failuref(
		"ifplugd: link-state monitoring of %q requires privileged netlink access not available in this "+
			"environment (capability-gated backend)", *iface)
}

// shellHook returns the production HookRunner: it executes the hook command
// with "/bin/sh -c", wiring the applet's stdout/stderr to the child.
func shellHook(stdio command.IO) HookRunner {
	return func(ctx context.Context, cmdline string) error {
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", cmdline)
		cmd.Stdout = stdio.Out
		cmd.Stderr = stdio.Err
		return cmd.Run()
	}
}

func verb(bringUp bool) string {
	if bringUp {
		return "up"
	}
	return "down"
}
