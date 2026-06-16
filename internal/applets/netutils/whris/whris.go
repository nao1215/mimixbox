// Package whris implements the whris applet: display management information
// (owner / AS number) for the IP addresses a domain resolves to.
//
// This is a clean-room reimplementation. The original morrigan whris was forked
// from harakeishi/whris (MIT License, Copyright (c) harakeishi); that
// attribution is preserved here per the MIT terms even though no source is
// copied verbatim.
package whris

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the whris applet.
type Command struct{}

// New returns a whris command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "whris" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show management information for a domain's IP addresses" }

// info holds the management information for one IP address.
type info struct {
	ip    string
	asn   string
	owner string
}

// resolver looks up the IP addresses of a host; tests replace it.
var resolver = func(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

// asnLookup returns the AS number and owner for an IP; tests replace it. The
// default queries Team Cymru's WHOIS IP-to-ASN service.
var asnLookup = cymruLookup

// Run executes whris.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "DOMAIN", stdio.Err).WithHelp(command.Help{
		Description: "Resolve DOMAIN to its IPv4 addresses and print the AS number and owning " +
			"organization for each, looked up through the Team Cymru WHOIS service.",
		Examples: []command.Example{
			{Command: "whris example.com", Explain: "Show the AS number and owner for example.com's IP addresses."},
		},
		ExitStatus: "0  the management information was printed.\n1  the domain could not be resolved or no domain was given.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	domains := fs.Args()
	if len(domains) != 1 {
		return command.Failuref("exactly one DOMAIN is required")
	}

	ips, err := resolver(domains[0])
	if err != nil {
		return command.Failuref("cannot resolve %q: %v", domains[0], err)
	}

	infos := collect(ips)
	for _, in := range infos {
		_, _ = fmt.Fprintf(stdio.Out, "%s\tAS%s\t%s\n", in.ip, in.asn, in.owner)
	}
	return nil
}

// collect gathers the management info for each IPv4 address, skipping lookups
// that fail so one bad IP does not abort the whole report.
func collect(ips []net.IP) []info {
	var infos []info
	for _, ip := range ips {
		v4 := ip.To4()
		if v4 == nil {
			continue
		}
		asn, owner, err := asnLookup(v4.String())
		if err != nil {
			asn, owner = "?", "lookup failed"
		}
		infos = append(infos, info{ip: v4.String(), asn: asn, owner: owner})
	}
	return infos
}

// cymruServer is the Team Cymru WHOIS endpoint; tests point it at a local stub.
var cymruServer = "whois.cymru.com:43"

// cymruLookup queries whois.cymru.com for the AS number and owner of ip.
func cymruLookup(ip string) (asn, owner string, err error) {
	conn, err := net.DialTimeout("tcp", cymruServer, 5*time.Second)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := fmt.Fprintf(conn, " -v %s\n", ip); err != nil {
		return "", "", err
	}
	return parseCymru(conn)
}

// parseCymru reads the Team Cymru "-v" response and returns the AS number and
// owner from the data row (the second non-header line).
func parseCymru(r interface{ Read([]byte) (int, error) }) (asn, owner string, err error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "AS") && strings.Contains(line, "|") {
			continue // header: "AS | IP | ..."
		}
		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}
		return strings.TrimSpace(fields[0]), strings.TrimSpace(fields[6]), nil
	}
	if err := sc.Err(); err != nil {
		return "", "", err
	}
	return "", "", fmt.Errorf("no data in WHOIS response")
}
