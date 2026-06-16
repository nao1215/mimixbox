// Package hostid implements the hostid applet: print the numeric identifier of
// the current host as a zero-padded 8-digit hexadecimal number, matching GNU
// coreutils' hostid (which prints gethostid() & 0xffffffff).
package hostid

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"

	"github.com/nao1215/mimixbox/internal/command"
)

// hostidFile is the path glibc's gethostid() consults first. It is a variable
// so tests can point it at a temporary file.
var hostidFile = "/etc/hostid"

// Command is the hostid applet.
type Command struct{}

// New returns a hostid command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "hostid" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	return "Print the numeric identifier (in hexadecimal) for the current host"
}

// Run executes hostid.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]", stdio.Err).WithHelp(command.Help{
		Description: "Print the numeric identifier of the current host as a zero-padded 8-digit hexadecimal number, " +
			"matching GNU coreutils' hostid.",
		Examples: []command.Example{
			{Command: "hostid", Explain: "Print the host identifier, for example 007f0101."},
		},
		ExitStatus: "0  the host identifier was printed successfully.",
	})

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	// POSIX says gethostid returns a 32-bit identifier; coreutils masks off any
	// sign extension and prints it zero-padded to 8 hex digits.
	_, _ = fmt.Fprintf(stdio.Out, "%08x\n", hostID()&0xffffffff)
	return nil
}

// hostID reproduces glibc's gethostid(): if /etc/hostid holds at least four
// bytes, those are the host id; otherwise it is derived from the host name's
// IPv4 address by swapping its two 16-bit halves.
func hostID() uint32 {
	if id, ok := hostIDFromFile(hostidFile); ok {
		return id
	}
	return hostIDFromHostname(os.Hostname, net.LookupIP)
}

// hostIDFromFile reads the first four bytes of path as a host-byte-order
// (little-endian on the supported platforms) unsigned 32-bit integer.
func hostIDFromFile(path string) (uint32, bool) {
	b, err := os.ReadFile(path) //nolint:gosec // /etc/hostid is a well-known system file
	if err != nil || len(b) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[:4]), true
}

// hostIDFromHostname derives the id from the first IPv4 address the host name
// resolves to, with the two 16-bit halves of the address swapped (the glibc
// fallback). It returns 0 when the name cannot be resolved to an IPv4 address,
// matching glibc/coreutils, which then print "00000000".
func hostIDFromHostname(hostname func() (string, error), lookup func(string) ([]net.IP, error)) uint32 {
	name, err := hostname()
	if err != nil {
		return 0
	}
	ips, err := lookup(name)
	if err != nil {
		return 0
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			s := binary.LittleEndian.Uint32(v4)
			return (s << 16) | (s >> 16)
		}
	}
	return 0
}
