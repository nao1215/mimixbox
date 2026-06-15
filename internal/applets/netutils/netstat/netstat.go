// Package netstat implements the netstat applet in read-only inspection mode: it
// prints network connections and listening sockets from an injectable data
// source. All data is supplied by a fixture in tests, so the command never reads
// the live kernel and stays hermetic.
package netstat

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/nao1215/mimixbox/internal/command"
)

// Socket is one network socket entry netstat can display.
type Socket struct {
	Proto   string // tcp, udp, tcp6, udp6
	Local   string // local address:port
	Foreign string // foreign address:port (or "*:*")
	State   string // e.g. LISTEN, ESTABLISHED (empty for UDP)
}

// Command is the netstat applet.
type Command struct{}

// New returns a netstat command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "netstat" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Show network connections and sockets (read-only)" }

// source supplies socket fixtures; tests replace it via SetSource.
var source = func() []Socket { return nil }

// SetSource installs fixture sockets for a test and returns a restore func.
func SetSource(sockets []Socket) (restore func()) {
	orig := source
	source = func() []Socket { return sockets }
	return func() { source = orig }
}

// Run executes netstat.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-t] [-u] [-l] [-a] [-n]", stdio.Err).WithHelp(command.Help{
		Description: "Print network connections and sockets in the traditional columnar format " +
			"(Proto, Local Address, Foreign Address, State). Filter by protocol with -t (TCP) and " +
			"-u (UDP); restrict to listening sockets with -l. With no protocol filter, both TCP and " +
			"UDP are shown. This slice renders from an injected data source and is read-only.",
		Examples: []command.Example{
			{Command: "netstat -tln", Explain: "Show listening TCP sockets, numerically."},
			{Command: "netstat -tu", Explain: "Show TCP and UDP connections."},
		},
		ExitStatus: "0  the socket table was printed.",
		Notes:      []string{"Process/PID columns (-p) and statistics (-s) are not implemented in this slice."},
	})
	tcp := fs.BoolP("tcp", "t", false, "show TCP sockets")
	udp := fs.BoolP("udp", "u", false, "show UDP sockets")
	listening := fs.BoolP("listening", "l", false, "show only listening sockets")
	_ = fs.BoolP("all", "a", false, "show all sockets (listening and non-listening)")
	_ = fs.BoolP("numeric", "n", false, "show numerical addresses (always on in this slice)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	wantTCP, wantUDP := *tcp, *udp
	if !wantTCP && !wantUDP {
		wantTCP, wantUDP = true, true
	}

	tw := tabwriter.NewWriter(stdio.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Active Internet connections")
	fmt.Fprintln(tw, "Proto\tLocal Address\tForeign Address\tState")
	for _, s := range source() {
		if !protoWanted(s.Proto, wantTCP, wantUDP) {
			continue
		}
		if *listening && s.State != "LISTEN" {
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Proto, s.Local, s.Foreign, s.State)
	}
	_ = tw.Flush()
	return nil
}

// protoWanted reports whether socket protocol p passes the TCP/UDP filter.
func protoWanted(p string, wantTCP, wantUDP bool) bool {
	switch p {
	case "tcp", "tcp6":
		return wantTCP
	case "udp", "udp6":
		return wantUDP
	default:
		return wantTCP || wantUDP
	}
}
