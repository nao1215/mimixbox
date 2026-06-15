// Package ifconfig implements the ifconfig applet in read-only inspection mode:
// it lists configured network interfaces and their addresses from an injectable
// data source. Configuration (bringing interfaces up/down, assigning addresses)
// is intentionally deferred and reported as a documented capability error rather
// than performed silently.
package ifconfig

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ifconfig applet.
type Command struct{}

// New returns an ifconfig command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ifconfig" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show network interface configuration (read-only)" }

// source supplies interface fixtures; tests replace it via SetSource.
var source = func() []ipcmd.Link { return nil }

// SetSource installs fixture interfaces for a test and returns a restore func.
func SetSource(links []ipcmd.Link) (restore func()) {
	orig := source
	source = func() []ipcmd.Link { return links }
	return func() { source = orig }
}

// Run executes ifconfig.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [INTERFACE]", stdio.Err).WithHelp(command.Help{
		Description: "Display the configuration of network interfaces. With no operand, the active " +
			"interfaces are shown; with an INTERFACE operand, only that interface is shown. This " +
			"slice is read-only: assigning addresses or changing interface state is intentionally " +
			"deferred and reported as an error so nothing is changed silently.",
		Examples: []command.Example{
			{Command: "ifconfig", Explain: "Show all active interfaces."},
			{Command: "ifconfig eth0", Explain: "Show one interface."},
		},
		ExitStatus: "0  the interface configuration was printed.\n" +
			"1  a configuration operation was requested, or the interface was not found.",
		Notes: []string{"Configuration (up/down, address assignment) is not implemented in this slice."},
	})
	all := fs.BoolP("all", "a", false, "display all interfaces, including down ones")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) > 1 {
		return command.Failuref("configuration is not implemented in this read-only slice")
	}
	want := ""
	if len(operands) == 1 {
		want = operands[0]
	}

	links := source()
	matched := false
	for _, l := range links {
		if want != "" && l.Name != want {
			continue
		}
		if !*all && want == "" && !isUp(l) {
			continue
		}
		matched = true
		writeInterface(stdio.Out, l)
	}
	if want != "" && !matched {
		return command.Failuref("interface %q not found", want)
	}
	return nil
}

// isUp reports whether a link carries the UP flag.
func isUp(l ipcmd.Link) bool {
	for _, f := range l.Flags {
		if f == "UP" {
			return true
		}
	}
	return false
}

// writeInterface renders one interface block in classic ifconfig style.
func writeInterface(w interface{ Write([]byte) (int, error) }, l ipcmd.Link) {
	fmt.Fprintf(w, "%s: flags=<%s>  mtu %d\n", l.Name, strings.Join(l.Flags, ","), l.MTU)
	for _, a := range l.Addrs {
		ip := a.CIDR
		if i := strings.IndexByte(ip, '/'); i >= 0 {
			ip = ip[:i]
		}
		if a.Family == "inet6" {
			fmt.Fprintf(w, "        inet6 %s  scopeid %s\n", ip, a.Scope)
		} else {
			fmt.Fprintf(w, "        inet %s  netmask %s\n", ip, netmaskOf(a.CIDR))
		}
	}
	if l.MAC != "" {
		fmt.Fprintf(w, "        ether %s\n", l.MAC)
	}
}

// netmaskOf returns the dotted IPv4 netmask for a CIDR string, or "" if the
// prefix is missing or not a valid IPv4 prefix length.
func netmaskOf(cidr string) string {
	i := strings.IndexByte(cidr, '/')
	if i < 0 {
		return ""
	}
	prefix, err := strconv.Atoi(cidr[i+1:])
	if err != nil || prefix < 0 || prefix > 32 {
		return ""
	}
	return net.IP(net.CIDRMask(prefix, 32)).String()
}
