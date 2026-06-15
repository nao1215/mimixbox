// Package ipcalc implements the ipcalc applet: calculate IP network parameters
// (netmask, network, broadcast, prefix, and host range) from an address and a
// netmask or prefix length. It is a pure-computation applet and never touches
// the network, so its output is fully deterministic and testable.
package ipcalc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the ipcalc applet.
type Command struct{}

// New returns an ipcalc command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ipcalc" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Calculate IP network parameters from an address" }

// result holds every value ipcalc can compute for one address/mask pair. Empty
// fields are simply not printed.
type result struct {
	address   string
	netmask   string
	network   string
	broadcast string
	prefix    int
	hostMin   string
	hostMax   string
	hostCount uint64
}

// Run executes ipcalc.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... ADDRESS[/PREFIX] [NETMASK]", stdio.Err).WithHelp(command.Help{
		Description: "Compute IPv4 network parameters from an ADDRESS and a netmask. The mask may be " +
			"given as a /PREFIX suffix on the address (e.g. 192.168.10.7/24), as a dotted NETMASK " +
			"operand (e.g. 255.255.255.0), or with -m. When no mask is given the classful default " +
			"mask for the address is used. With no display option, ipcalc prints all values in a " +
			"human-readable table; the -b/-n/-p/-m/--network/--broadcast/--hostrange options select " +
			"individual values in machine-readable KEY=VALUE form (one per line).",
		Examples: []command.Example{
			{Command: "ipcalc 192.168.10.7/24", Explain: "Print the full table for the /24 network."},
			{Command: "ipcalc 10.0.0.1 255.255.255.0", Explain: "Use a dotted netmask operand."},
			{Command: "ipcalc -b -n 172.16.5.9/20", Explain: "Print only BROADCAST and NETWORK lines."},
			{Command: "ipcalc -p 192.168.1.1 255.255.255.128", Explain: "Print the prefix length for a mask."},
		},
		ExitStatus: "0  the input parsed and the requested values were printed.\n" +
			"1  the address or netmask was invalid.",
		Notes: []string{
			"Only IPv4 is supported in this slice; IPv6 input fails deterministically.",
			"HOSTMIN/HOSTMAX/HOSTS for /31 and /32 networks follow the BusyBox convention " +
				"(/31 has two usable hosts, /32 has one).",
		},
	})
	broadcast := fs.BoolP("broadcast", "b", false, "show the broadcast address")
	network := fs.BoolP("network", "n", false, "show the network address")
	prefix := fs.BoolP("prefix", "p", false, "show the prefix length")
	netmask := fs.BoolP("netmask", "", false, "show the netmask")
	hostrange := fs.BoolP("hostrange", "", false, "show the usable host address range")
	maskOpt := fs.StringP("mask", "m", "", "netmask to use (dotted or prefix length)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) < 1 || len(operands) > 2 {
		return command.Failuref("expected ADDRESS[/PREFIX] and optional NETMASK")
	}

	maskOperand := ""
	if len(operands) == 2 {
		maskOperand = operands[1]
	}
	res, err := compute(operands[0], maskOperand, *maskOpt)
	if err != nil {
		return command.Failuref("%v", err)
	}

	selective := *broadcast || *network || *prefix || *netmask || *hostrange
	if selective {
		printSelective(stdio.Out, res, selectFlags{
			broadcast: *broadcast,
			network:   *network,
			prefix:    *prefix,
			netmask:   *netmask,
			hostrange: *hostrange,
		})
		return nil
	}
	printTable(stdio.Out, res)
	return nil
}

// selectFlags records which individual values the user asked for.
type selectFlags struct {
	broadcast bool
	network   bool
	prefix    bool
	netmask   bool
	hostrange bool
}

// compute parses addrSpec (ADDRESS or ADDRESS/PREFIX), reconciles it with an
// optional dotted/prefix maskOperand and the -m maskFlag, and returns every
// derived value. The precedence for the mask is: explicit /PREFIX on the
// address, then NETMASK operand, then -m, then the classful default.
func compute(addrSpec, maskOperand, maskFlag string) (result, error) {
	ipStr, prefixFromSpec, hasPrefix := splitPrefix(addrSpec)

	ip := net.ParseIP(ipStr)
	if ip == nil || ip.To4() == nil {
		return result{}, fmt.Errorf("invalid IPv4 address: %q", ipStr)
	}
	ip = ip.To4()

	mask, err := resolveMask(ip, prefixFromSpec, hasPrefix, maskOperand, maskFlag)
	if err != nil {
		return result{}, err
	}

	ones, _ := mask.Size()
	netIP := ip.Mask(mask)
	bcast := broadcastOf(netIP, mask)

	res := result{
		address:   ip.String(),
		netmask:   net.IP(mask).String(),
		network:   netIP.String(),
		broadcast: bcast.String(),
		prefix:    ones,
	}
	res.hostMin, res.hostMax, res.hostCount = hostRange(netIP, bcast, ones)
	return res, nil
}

// splitPrefix separates an "address/prefix" spec into the address and the
// prefix value. hasPrefix reports whether a "/n" suffix was present.
func splitPrefix(spec string) (ip string, prefix int, hasPrefix bool) {
	i := strings.IndexByte(spec, '/')
	if i < 0 {
		return spec, 0, false
	}
	p, err := strconv.Atoi(spec[i+1:])
	if err != nil {
		// Leave hasPrefix true so resolveMask reports the bad prefix.
		return spec[:i], -1, true
	}
	return spec[:i], p, true
}

// resolveMask picks the IPv4 mask to use following the documented precedence and
// validates whichever source it chose.
func resolveMask(ip net.IP, prefixFromSpec int, hasPrefix bool, maskOperand, maskFlag string) (net.IPMask, error) {
	switch {
	case hasPrefix:
		return maskFromPrefix(prefixFromSpec)
	case maskOperand != "":
		return maskFromString(maskOperand)
	case maskFlag != "":
		return maskFromString(maskFlag)
	default:
		return classfulMask(ip), nil
	}
}

// maskFromString accepts either a dotted-quad netmask or a bare prefix length.
func maskFromString(s string) (net.IPMask, error) {
	if !strings.Contains(s, ".") {
		p, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid netmask: %q", s)
		}
		return maskFromPrefix(p)
	}
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("invalid netmask: %q", s)
	}
	mask := net.IPMask(ip.To4())
	if ones, bits := mask.Size(); ones == 0 && bits == 0 {
		return nil, fmt.Errorf("non-contiguous netmask: %q", s)
	}
	return mask, nil
}

// maskFromPrefix builds a /n IPv4 mask, rejecting out-of-range prefix lengths.
func maskFromPrefix(p int) (net.IPMask, error) {
	if p < 0 || p > 32 {
		return nil, fmt.Errorf("invalid prefix length: %d (must be 0-32)", p)
	}
	return net.CIDRMask(p, 32), nil
}

// classfulMask returns the historical class A/B/C default mask for ip; anything
// outside those ranges (e.g. multicast) defaults to /24.
func classfulMask(ip net.IP) net.IPMask {
	switch first := ip[0]; {
	case first < 128:
		return net.CIDRMask(8, 32)
	case first < 192:
		return net.CIDRMask(16, 32)
	default:
		return net.CIDRMask(24, 32)
	}
}

// broadcastOf returns the broadcast address for a network address and mask.
func broadcastOf(netIP net.IP, mask net.IPMask) net.IP {
	b := make(net.IP, len(netIP))
	for i := range netIP {
		b[i] = netIP[i] | ^mask[i]
	}
	return b
}

// hostRange returns the first and last usable host addresses and the usable
// host count, applying the BusyBox special cases for /31 and /32.
func hostRange(netIP, bcast net.IP, prefix int) (min, max string, count uint64) {
	switch prefix {
	case 32:
		return netIP.String(), netIP.String(), 1
	case 31:
		return netIP.String(), bcast.String(), 2
	default:
		return incIP(netIP).String(), decIP(bcast).String(), (uint64(1) << uint(32-prefix)) - 2
	}
}

// incIP returns ip + 1 (IPv4).
func incIP(ip net.IP) net.IP {
	out := make(net.IP, len(ip))
	copy(out, ip)
	for i := len(out) - 1; i >= 0; i-- {
		out[i]++
		if out[i] != 0 {
			break
		}
	}
	return out
}

// decIP returns ip - 1 (IPv4).
func decIP(ip net.IP) net.IP {
	out := make(net.IP, len(ip))
	copy(out, ip)
	for i := len(out) - 1; i >= 0; i-- {
		if out[i] != 0 {
			out[i]--
			break
		}
		out[i] = 0xff
	}
	return out
}

// printTable writes the human-readable, aligned table.
func printTable(w interface{ Write([]byte) (int, error) }, r result) {
	fmt.Fprintf(w, "Address:   %s\n", r.address)
	fmt.Fprintf(w, "Netmask:   %s = %d\n", r.netmask, r.prefix)
	fmt.Fprintf(w, "Network:   %s/%d\n", r.network, r.prefix)
	fmt.Fprintf(w, "Broadcast: %s\n", r.broadcast)
	fmt.Fprintf(w, "HostMin:   %s\n", r.hostMin)
	fmt.Fprintf(w, "HostMax:   %s\n", r.hostMax)
	fmt.Fprintf(w, "Hosts:     %d\n", r.hostCount)
}

// printSelective writes only the requested KEY=VALUE lines, in a stable order.
func printSelective(w interface{ Write([]byte) (int, error) }, r result, sel selectFlags) {
	if sel.netmask {
		fmt.Fprintf(w, "NETMASK=%s\n", r.netmask)
	}
	if sel.prefix {
		fmt.Fprintf(w, "PREFIX=%d\n", r.prefix)
	}
	if sel.network {
		fmt.Fprintf(w, "NETWORK=%s\n", r.network)
	}
	if sel.broadcast {
		fmt.Fprintf(w, "BROADCAST=%s\n", r.broadcast)
	}
	if sel.hostrange {
		fmt.Fprintf(w, "HOSTMIN=%s\n", r.hostMin)
		fmt.Fprintf(w, "HOSTMAX=%s\n", r.hostMax)
		fmt.Fprintf(w, "HOSTS=%d\n", r.hostCount)
	}
}
