// Package dnsd implements the dnsd applet: a tiny authoritative DNS server that
// answers A-record queries from a hosts file.
//
// The wire-format encoding/decoding and the hosts-file parser are pure
// functions so they can be table-tested, and the foreground UDP server runs over
// loopback for hermetic integration tests.
package dnsd

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the dnsd applet.
type Command struct{}

// New returns a dnsd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "dnsd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Tiny authoritative DNS server for a hosts file" }

// Run executes dnsd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-p ADDR] -H HOSTSFILE", stdio.Err).WithHelp(command.Help{
		Description: "Answer DNS A-record queries from a hosts file. -f keeps dnsd in the foreground; -p " +
			"sets the UDP listen address (default 127.0.0.1:53); -H names the hosts file. Each hosts-file " +
			"line is 'IPV4 NAME' (BusyBox dnsd order); blank lines and '#' comments are ignored. Unknown " +
			"names get an NXDOMAIN reply. The server runs until its context is cancelled.",
		Examples: []command.Example{
			{Command: "dnsd -f -p 127.0.0.1:5353 -H hosts.txt", Explain: "Serve A records from hosts.txt on loopback port 5353."},
		},
		ExitStatus: "0  clean shutdown.\n1  bad arguments, missing hosts file, or bind error.",
		Notes: []string{
			"Only A (IPv4) records and the IN class are answered; other query types return NXDOMAIN.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("port", "p", "127.0.0.1:53", "UDP address to listen on (HOST:PORT)")
	hostsFile := fs.StringP("hosts", "H", "", "hosts file mapping names to IPv4 addresses")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}
	if *hostsFile == "" {
		return command.Failuref("a hosts file is required (-H)")
	}

	f, err := os.Open(*hostsFile)
	if err != nil {
		return command.Failuref("cannot open hosts file %q: %v", *hostsFile, err)
	}
	defer func() { _ = f.Close() }()
	zone, err := ParseHosts(f)
	if err != nil {
		return command.Failuref("%s: %v", *hostsFile, err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", *addr)
	if err != nil {
		return command.Failuref("resolve %s: %v", *addr, err)
	}
	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "dnsd: serving %d names on %s\n", len(zone), pc.LocalAddr().String())
	return Serve(ctx, pc, zone)
}

// ParseHosts reads "IPV4 NAME" lines into a name->IP map (names lowercased).
func ParseHosts(r io.Reader) (map[string]net.IP, error) {
	zone := make(map[string]net.IP)
	sc := bufio.NewScanner(r)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		f := strings.Fields(text)
		if len(f) < 2 {
			return nil, fmt.Errorf("line %d: expected 'IPV4 NAME', got %q", line, text)
		}
		ip := net.ParseIP(f[0]).To4()
		if ip == nil {
			return nil, fmt.Errorf("line %d: invalid IPv4 address %q", line, f[0])
		}
		zone[strings.ToLower(strings.TrimSuffix(f[1], "."))] = ip
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return zone, nil
}

// Serve runs the UDP DNS loop on pc until ctx is cancelled.
func Serve(ctx context.Context, pc *net.UDPConn, zone map[string]net.IP) error {
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()
	buf := make([]byte, 512)
	for {
		n, raddr, err := pc.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return command.Failuref("read: %v", err)
		}
		reply, err := BuildResponse(buf[:n], zone)
		if err != nil {
			continue // ignore malformed queries
		}
		_, _ = pc.WriteToUDP(reply, raddr)
	}
}

// errMalformed marks a query that cannot be parsed.
var errMalformed = errors.New("malformed DNS query")

// BuildResponse parses a DNS query and builds an A-record response. A known name
// yields an answer; an unknown name yields NXDOMAIN (rcode 3). Only A/IN queries
// produce answers.
func BuildResponse(query []byte, zone map[string]net.IP) ([]byte, error) {
	name, qtype, qclass, qend, err := parseQuestion(query)
	if err != nil {
		return nil, err
	}

	id := binary.BigEndian.Uint16(query[0:2])
	resp := make([]byte, 0, len(query)+16)
	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], id)
	// QR=1, Opcode=0, AA=1, RD copied from query, RA=0.
	flags := uint16(0x8400)
	flags |= uint16(query[2]&0x01) << 8 // preserve RD bit
	binary.BigEndian.PutUint16(header[2:4], flags)
	binary.BigEndian.PutUint16(header[4:6], 1) // QDCOUNT

	ip, ok := zone[strings.ToLower(name)]
	answer := qtype == 1 && qclass == 1 && ok
	if answer {
		binary.BigEndian.PutUint16(header[6:8], 1) // ANCOUNT
	} else {
		binary.BigEndian.PutUint16(header[6:8], 0)
		header[3] |= 0x03 // NXDOMAIN rcode
	}

	resp = append(resp, header...)
	resp = append(resp, query[12:qend]...) // echo the question

	if answer {
		ans := make([]byte, 0, 16)
		ans = append(ans, 0xc0, 0x0c)                 // name pointer to offset 12
		ans = append(ans, 0x00, 0x01)                 // type A
		ans = append(ans, 0x00, 0x01)                 // class IN
		ans = append(ans, 0x00, 0x00, 0x00, 0x3c)     // TTL 60
		ans = append(ans, 0x00, 0x04)                 // RDLENGTH 4
		ans = append(ans, ip.To4()...)                // RDATA
		resp = append(resp, ans...)
	}
	return resp, nil
}

// parseQuestion decodes the single question in a DNS query, returning the name,
// qtype, qclass, and the byte offset just past the question section.
func parseQuestion(q []byte) (name string, qtype, qclass uint16, end int, err error) {
	if len(q) < 12 {
		return "", 0, 0, 0, errMalformed
	}
	if binary.BigEndian.Uint16(q[4:6]) < 1 {
		return "", 0, 0, 0, errMalformed
	}
	var labels []string
	i := 12
	for {
		if i >= len(q) {
			return "", 0, 0, 0, errMalformed
		}
		l := int(q[i])
		i++
		if l == 0 {
			break
		}
		if l > 63 || i+l > len(q) {
			return "", 0, 0, 0, errMalformed
		}
		labels = append(labels, string(q[i:i+l]))
		i += l
	}
	if i+4 > len(q) {
		return "", 0, 0, 0, errMalformed
	}
	qtype = binary.BigEndian.Uint16(q[i : i+2])
	qclass = binary.BigEndian.Uint16(q[i+2 : i+4])
	return strings.Join(labels, "."), qtype, qclass, i + 4, nil
}
