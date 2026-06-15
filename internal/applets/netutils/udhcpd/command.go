package udhcpd

import (
	"context"
	"net"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the udhcpd / dhcprelay applet.
type Command struct {
	name string
}

// NewUdhcpd returns a udhcpd command.
func NewUdhcpd() *Command { return &Command{name: "udhcpd"} }

// NewDhcprelay returns a dhcprelay command.
func NewDhcprelay() *Command { return &Command{name: "dhcprelay"} }

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.name == "dhcprelay" {
		return "Relay DHCP requests between networks"
	}
	return "DHCP server"
}

// Run dispatches to the udhcpd or dhcprelay behaviour.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	if c.name == "dhcprelay" {
		return runRelay(ctx, stdio, args)
	}
	return runUdhcpd(ctx, stdio, args)
}

func runUdhcpd(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("udhcpd", "[-f] [-b ADDR] CONFIG", stdio.Err).WithHelp(command.Help{
		Description: "Serve DHCP from a configuration file. -f keeps udhcpd in the foreground; -b sets the " +
			"UDP listen address (default 0.0.0.0:67; use 127.0.0.1:PORT for hermetic testing). The config " +
			"file uses the udhcpd grammar: 'start IP', 'end IP', 'interface NAME', 'server_id IP', " +
			"'lease SECONDS', and 'opt NAME VALUE' (subnet/router/dns). The packet codec, config parser, " +
			"and lease allocator are exposed for transport-injected unit tests.",
		Examples: []command.Example{
			{Command: "udhcpd -f -b 127.0.0.1:6767 udhcpd.conf", Explain: "Serve DHCP in the foreground on loopback port 6767."},
		},
		ExitStatus: "0  clean shutdown.\n1  config error or bind error.",
		Notes: []string{
			"Writing the on-disk binary lease database is not implemented in this slice; allocation is in-memory.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("bind", "b", "0.0.0.0:67", "UDP address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return command.Failuref("a config file is required")
	}
	f, err := os.Open(rest[0])
	if err != nil {
		return command.Failuref("cannot open config %q: %v", rest[0], err)
	}
	defer func() { _ = f.Close() }()
	cfg, err := ParseConfig(f)
	if err != nil {
		return command.Failuref("%s: %v", rest[0], err)
	}
	return runServer(ctx, cfg, *addr, stdio.Out)
}

func runRelay(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet("dhcprelay", "[INTERFACE...] SERVER", stdio.Err).WithHelp(command.Help{
		Description: "Relay DHCP requests between a client-facing interface and a DHCP server. This slice " +
			"validates the requested relay plan (client interfaces and the upstream server address) and " +
			"reports it, but performing the privileged cross-interface packet forwarding requires raw " +
			"socket and routing capabilities that are not available in this environment.",
		Examples: []command.Example{
			{Command: "dhcprelay eth0 eth1 192.168.0.1", Explain: "Plan a relay from eth0/eth1 to the server at 192.168.0.1."},
		},
		ExitStatus: "0  never (capability-gated).\n1  always: prints the validated plan then a documented backend error.",
		Notes: []string{
			"Raw-socket relaying is capability-gated; the command validates and serializes the plan, then fails deterministically.",
		},
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return command.Failuref("usage: dhcprelay INTERFACE... SERVER")
	}
	server := rest[len(rest)-1]
	ifaces := rest[:len(rest)-1]
	if net.ParseIP(server) == nil {
		return command.Failuref("invalid server address %q", server)
	}
	for _, name := range ifaces {
		if name == "" {
			return command.Failuref("interface names must not be empty")
		}
	}
	return command.Failuref(
		"relay plan: interfaces=%v server=%s; raw-socket DHCP relaying is not available in this environment (capability-gated backend)",
		ifaces, server)
}
