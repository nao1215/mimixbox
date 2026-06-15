package ipcmd

import (
	"fmt"
	"io"
	"strings"
)

// NetData is the injectable snapshot the read-only ip commands render. The
// default source returns an empty snapshot so go test stays hermetic; tests
// install fixtures via SetSource.
type NetData struct {
	Links      []Link
	Routes     []Route
	Neighbours []Neighbour
	Rules      []Rule
}

// Link is one network device.
type Link struct {
	Index int
	Name  string
	Flags []string // e.g. UP, LOWER_UP, BROADCAST
	MTU   int
	MAC   string
	State string // e.g. UP, DOWN, UNKNOWN
	Addrs []Addr // addresses for "ip addr show"
}

// Addr is one IP address bound to a device.
type Addr struct {
	Family string // "inet" or "inet6"
	CIDR   string // e.g. 192.168.1.10/24
	Scope  string // e.g. global, host, link
}

// Route is one routing-table entry.
type Route struct {
	Dest   string // "default" or a CIDR
	Via    string // gateway, may be empty
	Dev    string
	Proto  string
	Scope  string
	Src    string
	Metric int
}

// Neighbour is one ARP/neighbour-table entry.
type Neighbour struct {
	IP    string
	Dev   string
	MAC   string
	State string // e.g. REACHABLE, STALE, FAILED
}

// Rule is one routing policy rule.
type Rule struct {
	Priority int
	Selector string // e.g. "from all"
	Action   string // e.g. "lookup main"
}

// defaultSource returns an empty snapshot, keeping the live kernel out of tests.
func defaultSource() NetData { return NetData{} }

// SetSource installs a fixture data source for the duration of a test and
// returns a restore function. It is exported so sibling applets (ifconfig,
// route, netstat, arp) can reuse the same fixtures.
func SetSource(d NetData) (restore func()) {
	orig := source
	source = func() NetData { return d }
	return func() { source = orig }
}

// writeLinks renders "ip link show".
func writeLinks(w io.Writer, links []Link, dev string) {
	for _, l := range links {
		if dev != "" && l.Name != dev {
			continue
		}
		fmt.Fprintf(w, "%d: %s: <%s> mtu %d state %s\n",
			l.Index, l.Name, strings.Join(l.Flags, ","), l.MTU, l.State)
		if l.MAC != "" {
			fmt.Fprintf(w, "    link/ether %s\n", l.MAC)
		}
	}
}

// writeAddrs renders "ip addr show".
func writeAddrs(w io.Writer, links []Link, dev string) {
	for _, l := range links {
		if dev != "" && l.Name != dev {
			continue
		}
		fmt.Fprintf(w, "%d: %s: <%s> mtu %d state %s\n",
			l.Index, l.Name, strings.Join(l.Flags, ","), l.MTU, l.State)
		if l.MAC != "" {
			fmt.Fprintf(w, "    link/ether %s\n", l.MAC)
		}
		for _, a := range l.Addrs {
			fmt.Fprintf(w, "    %s %s scope %s\n", a.Family, a.CIDR, a.Scope)
		}
	}
}

// writeRoutes renders "ip route show".
func writeRoutes(w io.Writer, routes []Route) {
	for _, r := range routes {
		var b strings.Builder
		b.WriteString(r.Dest)
		if r.Via != "" {
			fmt.Fprintf(&b, " via %s", r.Via)
		}
		if r.Dev != "" {
			fmt.Fprintf(&b, " dev %s", r.Dev)
		}
		if r.Proto != "" {
			fmt.Fprintf(&b, " proto %s", r.Proto)
		}
		if r.Scope != "" {
			fmt.Fprintf(&b, " scope %s", r.Scope)
		}
		if r.Src != "" {
			fmt.Fprintf(&b, " src %s", r.Src)
		}
		if r.Metric > 0 {
			fmt.Fprintf(&b, " metric %d", r.Metric)
		}
		fmt.Fprintln(w, b.String())
	}
}

// writeNeighbours renders "ip neigh show".
func writeNeighbours(w io.Writer, ns []Neighbour) {
	for _, n := range ns {
		if n.MAC != "" {
			fmt.Fprintf(w, "%s dev %s lladdr %s %s\n", n.IP, n.Dev, n.MAC, n.State)
		} else {
			fmt.Fprintf(w, "%s dev %s %s\n", n.IP, n.Dev, n.State)
		}
	}
}

// writeRules renders "ip rule show".
func writeRules(w io.Writer, rules []Rule) {
	for _, r := range rules {
		fmt.Fprintf(w, "%d:\t%s %s\n", r.Priority, r.Selector, r.Action)
	}
}
