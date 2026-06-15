// Package dnsdomainname implements the dnsdomainname applet: print the DNS
// domain part of the system's fully qualified host name. The hostname source and
// the reverse resolver are injectable so tests never depend on the host's real
// name or the public internet.
package dnsdomainname

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the dnsdomainname applet.
type Command struct{}

// New returns a dnsdomainname command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dnsdomainname" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show the DNS domain name" }

// hostnameFn returns the local host name; tests replace it.
var hostnameFn = os.Hostname

// lookupAddrFn resolves an IP to its host names; tests replace it.
var lookupAddrFn = net.LookupAddr

// lookupHostFn resolves a host name to its IP addresses; tests replace it.
var lookupHostFn = net.LookupHost

// Run executes dnsdomainname.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "", stdio.Err).WithHelp(command.Help{
		Description: "Print the DNS domain name: the portion of the fully qualified domain name (FQDN) " +
			"after the first label. The FQDN is obtained by resolving the local host name to an " +
			"address and resolving that address back to a canonical name. If the resolved name has " +
			"no dot (no domain part), nothing is printed and the exit status is still 0.",
		Examples: []command.Example{
			{Command: "dnsdomainname", Explain: "Print the local DNS domain (e.g. example.com)."},
		},
		ExitStatus: "0  the domain (possibly empty) was determined.\n" +
			"1  the host name could not be resolved to an FQDN.",
	})
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if len(fs.Args()) != 0 {
		return command.Failuref("dnsdomainname takes no operands")
	}

	fqdn, err := resolveFQDN()
	if err != nil {
		return command.Failuref("%v", err)
	}
	if domain := domainOf(fqdn); domain != "" {
		_, _ = fmt.Fprintln(stdio.Out, domain)
	}
	return nil
}

// resolveFQDN derives the fully qualified domain name from the local host name
// using the injected resolvers, mirroring the classic hostname --fqdn path.
func resolveFQDN() (string, error) {
	host, err := hostnameFn()
	if err != nil {
		return "", fmt.Errorf("cannot determine host name: %v", err)
	}
	if strings.Contains(host, ".") {
		return host, nil
	}

	addrs, err := lookupHostFn(host)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("cannot resolve host %q to an address", host)
	}
	names, err := lookupAddrFn(addrs[0])
	if err != nil || len(names) == 0 {
		return "", fmt.Errorf("cannot resolve address %q to a name", addrs[0])
	}
	return strings.TrimSuffix(names[0], "."), nil
}

// domainOf returns the domain portion (everything after the first label) of an
// FQDN, or "" when there is no domain part.
func domainOf(fqdn string) string {
	if i := strings.IndexByte(fqdn, '.'); i >= 0 {
		return fqdn[i+1:]
	}
	return ""
}
