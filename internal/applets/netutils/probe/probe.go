// Package probe implements the privileged-transport networking applets that
// share one shape: parse arguments and a target, then hand off to a transport
// that normally needs raw sockets / CAP_NET_RAW (traceroute, traceroute6, ping6,
// arping). The argument parsing and formatting are fully testable; the transport
// is an injectable function that, by default, reports a deterministic capability
// error instead of attempting a privileged operation. This satisfies the
// "never a silent no-op" rule: the first slice fails clearly and explains why.
package probe

import (
	"context"
	"fmt"
	"net"

	"github.com/nao1215/mimixbox/internal/command"
)

// kind identifies which probe applet a Command is.
type kind int

const (
	kindTraceroute kind = iota
	kindTraceroute6
	kindPing6
	kindArping
)

// spec describes the static properties of one probe applet.
type spec struct {
	name     string
	synopsis string
	usage    string
	want     ipWant // address family the target must parse as
	desc     string
}

// ipWant constrains the target address family.
type ipWant int

const (
	anyIP ipWant = iota
	ipv4Only
	ipv6Only
)

var specs = map[kind]spec{
	kindTraceroute: {
		name: "traceroute", synopsis: "Trace the route packets take to a host (IPv4)",
		usage: "HOST", want: ipv4Only,
		desc: "Print the route (sequence of hops) IPv4 packets take to HOST by sending probes with " +
			"increasing TTL.",
	},
	kindTraceroute6: {
		name: "traceroute6", synopsis: "Trace the route packets take to a host (IPv6)",
		usage: "HOST", want: ipv6Only,
		desc: "Print the route (sequence of hops) IPv6 packets take to HOST by sending probes with " +
			"increasing hop limit.",
	},
	kindPing6: {
		name: "ping6", synopsis: "Send ICMPv6 ECHO_REQUEST to a host", usage: "HOST", want: ipv6Only,
		desc: "Send ICMPv6 echo requests to HOST and report the replies.",
	},
	kindArping: {
		name: "arping", synopsis: "Probe a host on the local network with ARP requests",
		usage: "[-I INTERFACE] HOST", want: ipv4Only,
		desc: "Send ARP requests for HOST on the local link and report replies.",
	},
}

// Command is one probe applet.
type Command struct{ kind kind }

// NewTraceroute returns the traceroute applet.
func NewTraceroute() *Command { return &Command{kind: kindTraceroute} }

// NewTraceroute6 returns the traceroute6 applet.
func NewTraceroute6() *Command { return &Command{kind: kindTraceroute6} }

// NewPing6 returns the ping6 applet.
func NewPing6() *Command { return &Command{kind: kindPing6} }

// NewArping returns the arping applet.
func NewArping() *Command { return &Command{kind: kindArping} }

// Name returns the command name.
func (c *Command) Name() string { return specs[c.kind].name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return specs[c.kind].synopsis }

// Target is the parsed, validated probe target handed to a transport.
type Target struct {
	Host      string
	Interface string
}

// transport performs the privileged probe. The default reports a capability
// error; tests replace it to exercise the success path. Returning an error here
// is what keeps the first slice honest instead of a silent no-op.
var transport = func(_ context.Context, name string, _ Target) (string, error) {
	return "", fmt.Errorf("%s requires raw socket access (CAP_NET_RAW) which is not available; "+
		"the privileged transport is intentionally deferred in this slice", name)
}

// Run executes a probe applet.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	s := specs[c.kind]
	fs := command.NewFlagSet(c.Name(), s.usage, stdio.Err).WithHelp(command.Help{
		Description: s.desc + " Argument parsing and target validation are implemented and tested; the " +
			"actual probe needs raw-socket privileges (CAP_NET_RAW) and is intentionally deferred " +
			"in this slice. When the privileged transport is unavailable the command fails with a " +
			"clear capability error rather than doing nothing.",
		Examples: []command.Example{
			{Command: s.name + " example.test", Explain: "Probe a host (requires raw-socket capability)."},
		},
		ExitStatus: "0  the probe completed.\n" +
			"1  bad arguments, an unresolvable/ wrong-family target, or the transport is unavailable.",
		Notes: []string{"The privileged transport is not implemented; this slice validates input and reports a capability error."},
	})
	iface := fs.StringP("interface", "I", "", "probe via this network interface")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) != 1 {
		return command.Failuref("exactly one HOST is required")
	}
	target, err := c.parseTarget(operands[0], *iface)
	if err != nil {
		return command.Failuref("%v", err)
	}

	out, err := transport(ctx, c.Name(), target)
	if err != nil {
		return command.Failuref("%v", err)
	}
	_, _ = fmt.Fprint(stdio.Out, out)
	return nil
}

// parseTarget validates host against the applet's required address family. A
// literal IP is checked directly; a name is accepted as-is (resolution happens
// in the transport) so parsing stays hermetic.
func (c *Command) parseTarget(host, iface string) (Target, error) {
	want := specs[c.kind].want
	if ip := net.ParseIP(host); ip != nil {
		switch {
		case want == ipv4Only && ip.To4() == nil:
			return Target{}, fmt.Errorf("%s requires an IPv4 target, got %q", c.Name(), host)
		case want == ipv6Only && ip.To4() != nil:
			return Target{}, fmt.Errorf("%s requires an IPv6 target, got %q", c.Name(), host)
		}
	}
	return Target{Host: host, Interface: iface}, nil
}
