// Package nslookup implements the nslookup applet: resolve a NAME to its
// addresses (or an address to its names) using an injectable resolver, with an
// optional SERVER operand. Because the resolver is injected, the command's tests
// never depend on the public internet.
package nslookup

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the nslookup applet.
type Command struct{}

// New returns an nslookup command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "nslookup" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Query the DNS for a name or address" }

// Resolver abstracts the DNS operations nslookup needs so tests can inject a
// fixture that answers from memory.
type Resolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}

// newResolver builds the Resolver for a given DNS server. When server is empty
// the host's default resolver is used; otherwise queries are directed at
// server:53. Tests replace this to avoid real network access.
var newResolver = func(server string) Resolver {
	if server == "" {
		return stdResolver{net.DefaultResolver}
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, net.JoinHostPort(server, "53"))
		},
	}
	return stdResolver{r}
}

// stdResolver adapts *net.Resolver to the Resolver interface.
type stdResolver struct{ r *net.Resolver }

func (s stdResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return s.r.LookupHost(ctx, host)
}

func (s stdResolver) LookupAddr(ctx context.Context, addr string) ([]string, error) {
	return s.r.LookupAddr(ctx, addr)
}

// Run executes nslookup.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "NAME [SERVER]", stdio.Err).WithHelp(command.Help{
		Description: "Look up the DNS records for NAME. If NAME is an IPv4/IPv6 address, a reverse " +
			"(PTR) lookup is performed; otherwise a forward (A/AAAA) lookup is performed. An optional " +
			"SERVER operand selects the DNS server to query (port 53); without it the system default " +
			"resolver is used. The SERVER override exists so tests and scripts can point nslookup at " +
			"a local resolver instead of the public internet.",
		Examples: []command.Example{
			{Command: "nslookup example.test", Explain: "Forward lookup via the default resolver."},
			{Command: "nslookup example.test 127.0.0.1", Explain: "Query a specific server."},
			{Command: "nslookup 192.0.2.1", Explain: "Reverse (PTR) lookup."},
		},
		ExitStatus: "0  at least one record was found.\n" +
			"1  bad arguments or the lookup failed.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) < 1 || len(operands) > 2 {
		return command.Failuref("usage: nslookup NAME [SERVER]")
	}
	name := operands[0]
	server := ""
	if len(operands) == 2 {
		server = operands[1]
	}

	res := newResolver(server)
	if server != "" {
		fmt.Fprintf(stdio.Out, "Server:\t\t%s\n", server)
		fmt.Fprintf(stdio.Out, "Address:\t%s#53\n\n", server)
	}

	if ip := net.ParseIP(name); ip != nil {
		return reverse(ctx, stdio, res, name)
	}
	return forward(ctx, stdio, res, name)
}

// forward prints the A/AAAA records for name.
func forward(ctx context.Context, stdio command.IO, res Resolver, name string) error {
	addrs, err := res.LookupHost(ctx, name)
	if err != nil {
		return command.Failuref("can't resolve %q: %v", name, err)
	}
	if len(addrs) == 0 {
		return command.Failuref("can't resolve %q: no addresses", name)
	}
	fmt.Fprintf(stdio.Out, "Name:\t%s\n", name)
	for _, a := range addrs {
		fmt.Fprintf(stdio.Out, "Address: %s\n", a)
	}
	return nil
}

// reverse prints the PTR records for an address.
func reverse(ctx context.Context, stdio command.IO, res Resolver, addr string) error {
	names, err := res.LookupAddr(ctx, addr)
	if err != nil {
		return command.Failuref("can't resolve %q: %v", addr, err)
	}
	if len(names) == 0 {
		return command.Failuref("can't resolve %q: no names", addr)
	}
	for _, n := range names {
		fmt.Fprintf(stdio.Out, "%s\tname = %s\n", addr, strings.TrimSuffix(n, "."))
	}
	return nil
}
