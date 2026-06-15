// Package route implements the route applet in read-only inspection mode: it
// prints the kernel IPv4 routing table from an injectable data source in the
// classic "route -n" columnar format. Adding and deleting routes is
// intentionally deferred and reported as a documented capability error.
package route

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the route applet.
type Command struct{}

// New returns a route command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "route" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the IP routing table (read-only)" }

// source supplies route fixtures; tests replace it via SetSource.
var source = func() []ipcmd.Route { return nil }

// SetSource installs fixture routes for a test and returns a restore func.
func SetSource(routes []ipcmd.Route) (restore func()) {
	orig := source
	source = func() []ipcmd.Route { return routes }
	return func() { source = orig }
}

// Run executes route.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-n] [-e]", stdio.Err).WithHelp(command.Help{
		Description: "Print the kernel IPv4 routing table in the traditional columnar format " +
			"(Destination, Gateway, Genmask, Flags, Metric, Ref, Use, Iface). With -n, addresses " +
			"are shown numerically and never resolved. This slice is read-only: 'route add' and " +
			"'route del' are intentionally deferred and reported as an error.",
		Examples: []command.Example{
			{Command: "route -n", Explain: "Print the routing table numerically."},
		},
		ExitStatus: "0  the routing table was printed.\n" +
			"1  an add/del operation was requested.",
		Notes: []string{"Adding or deleting routes is not implemented in this slice."},
	})
	_ = fs.BoolP("numeric", "n", false, "show numerical addresses (always on in this slice)")
	_ = fs.BoolP("extend", "e", false, "use a longer output format")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) > 0 {
		switch operands[0] {
		case "add", "del", "delete":
			return command.Failuref("route %s is not implemented in this read-only slice", operands[0])
		default:
			return command.Failuref("unexpected operand %q", operands[0])
		}
	}

	writeTable(stdio.Out, source())
	return nil
}

// writeTable renders the routing table in "route -n" style.
func writeTable(w interface{ Write([]byte) (int, error) }, routes []ipcmd.Route) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Kernel IP routing table")
	fmt.Fprintln(tw, "Destination\tGateway\tGenmask\tFlags\tMetric\tRef\tUse\tIface")
	for _, r := range routes {
		dest, genmask := destAndMask(r.Dest)
		gw := r.Via
		if gw == "" {
			gw = "0.0.0.0"
		}
		flags := "U"
		if r.Via != "" {
			flags = "UG"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t0\t0\t%s\n",
			dest, gw, genmask, flags, r.Metric, r.Dev)
	}
	_ = tw.Flush()
}

// destAndMask splits a route destination ("default" or CIDR) into the numeric
// destination network and its dotted genmask.
func destAndMask(dest string) (network, genmask string) {
	if dest == "default" {
		return "0.0.0.0", "0.0.0.0"
	}
	if i := strings.IndexByte(dest, '/'); i >= 0 {
		net4 := dest[:i]
		prefix, err := strconv.Atoi(dest[i+1:])
		if err == nil && prefix >= 0 && prefix <= 32 {
			return net4, net.IP(net.CIDRMask(prefix, 32)).String()
		}
		return net4, ""
	}
	return dest, "255.255.255.255"
}
