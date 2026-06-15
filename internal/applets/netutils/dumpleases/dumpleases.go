// Package dumpleases implements the dumpleases applet: display the leases a DHCP
// server (udhcpd) has handed out.
//
// BusyBox dumpleases reads a binary lease database. Parsing that host-written
// binary format is not portable enough to test hermetically, so this slice
// reads a documented, deterministic text lease file (one lease per line). The
// parser is pure and table-tested; the applet only formats its output.
package dumpleases

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the dumpleases applet.
type Command struct{}

// New returns a dumpleases command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dumpleases" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Display DHCP server leases" }

// Lease is a single DHCP lease record.
type Lease struct {
	MAC      net.HardwareAddr
	IP       net.IP
	Hostname string
	Expires  time.Time // zero means "never" / static
}

// Run executes dumpleases.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-a] [-r] [-f LEASEFILE]", stdio.Err).WithHelp(command.Help{
		Description: "Read a DHCP lease file and print one lease per line as MAC address, IP address, " +
			"host name, and expiry. -f selects the lease file (default: /var/lib/misc/udhcpd.leases). " +
			"-a prints absolute expiry timestamps; the default prints the remaining time. -r prints the " +
			"remaining time explicitly. The lease file is a text file: each non-empty, non-'#' line holds " +
			"'MAC IP HOSTNAME EXPIRY', where EXPIRY is a Unix timestamp (seconds) or 0 for a static lease.",
		Examples: []command.Example{
			{Command: "dumpleases -f leases.db", Explain: "Print leases from leases.db with remaining time."},
			{Command: "dumpleases -a -f leases.db", Explain: "Print leases with absolute expiry timestamps."},
		},
		ExitStatus: "0  success.\n1  the lease file cannot be read or is malformed.",
		Notes: []string{
			"This slice reads a documented text lease format; BusyBox's binary lease database is not parsed.",
		},
	})
	absolute := fs.BoolP("absolute", "a", false, "show expiry as an absolute timestamp")
	_ = fs.BoolP("remaining", "r", false, "show remaining time until expiry (default)")
	file := fs.StringP("file", "f", "/var/lib/misc/udhcpd.leases", "lease file to read")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// A positional lease file (BusyBox accepts `dumpleases leases.db`) overrides -f.
	path := *file
	if rest := fs.Args(); len(rest) > 0 {
		path = rest[0]
	}

	f, err := os.Open(path)
	if err != nil {
		return command.Failuref("cannot open lease file %q: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	leases, err := ParseLeases(f)
	if err != nil {
		return command.Failuref("%s: %v", path, err)
	}

	writeLeases(stdio.Out, leases, *absolute, time.Now())
	return nil
}

// ParseLeases reads the text lease format from r. Lines are "MAC IP HOSTNAME
// EXPIRY"; blank lines and lines beginning with '#' are ignored.
func ParseLeases(r io.Reader) ([]Lease, error) {
	var leases []Lease
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) < 4 {
			return nil, fmt.Errorf("line %d: expected 'MAC IP HOSTNAME EXPIRY', got %q", line, text)
		}
		mac, err := net.ParseMAC(fields[0])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid MAC %q: %v", line, fields[0], err)
		}
		ip := net.ParseIP(fields[1])
		if ip == nil {
			return nil, fmt.Errorf("line %d: invalid IP %q", line, fields[1])
		}
		expSec, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid expiry %q: %v", line, fields[3], err)
		}
		l := Lease{MAC: mac, IP: ip, Hostname: fields[2]}
		if expSec > 0 {
			l.Expires = time.Unix(expSec, 0).UTC()
		}
		leases = append(leases, l)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return leases, nil
}

// writeLeases prints leases as an aligned table relative to now.
func writeLeases(w io.Writer, leases []Lease, absolute bool, now time.Time) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "Mac Address\tIP Address\tHost Name\tExpires")
	for _, l := range leases {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			l.MAC.String(), l.IP.String(), hostnameOrDash(l.Hostname), formatExpiry(l, absolute, now))
	}
	_ = tw.Flush()
}

func hostnameOrDash(h string) string {
	if h == "" || h == "*" {
		return "*"
	}
	return h
}

// formatExpiry renders the lease expiry: "never" for static leases, an absolute
// UTC timestamp when absolute is set, otherwise the remaining duration.
func formatExpiry(l Lease, absolute bool, now time.Time) string {
	if l.Expires.IsZero() {
		return "never"
	}
	if absolute {
		return l.Expires.Format(time.RFC3339)
	}
	d := l.Expires.Sub(now)
	if d <= 0 {
		return "expired"
	}
	return formatDuration(d)
}

// formatDuration renders d as "Dd HH:MM:SS"-style remaining time.
func formatDuration(d time.Duration) string {
	total := int64(d.Round(time.Second).Seconds())
	days := total / 86400
	total %= 86400
	h := total / 3600
	total %= 3600
	m := total / 60
	s := total % 60
	if days > 0 {
		return fmt.Sprintf("%d days %02d:%02d:%02d", days, h, m, s)
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
