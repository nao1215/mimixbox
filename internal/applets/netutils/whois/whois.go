// Package whois implements the whois applet: query a WHOIS server for a domain
// or IP and print the raw response. The transport is injectable and the server
// is configurable with -h, so tests run against a loopback stub and never touch
// the public internet.
package whois

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// Command is the whois applet.
type Command struct{}

// New returns a whois command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "whois" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Query a WHOIS server for a domain or IP" }

// defaultServer is the WHOIS server used when -h is not given.
const defaultServer = "whois.iana.org"

// query sends a WHOIS request for object to server:43 and returns the raw
// response. Tests replace it with a stub that answers from memory.
var query = tcpQuery

// Run executes whois.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-h SERVER] OBJECT", stdio.Err).WithHelp(command.Help{
		Description: "Query a WHOIS server for registration information about a domain name or IP " +
			"address OBJECT and print the server's raw response. The -h option selects the WHOIS " +
			"server (default " + defaultServer + ", port 43). The server override exists so tests " +
			"and scripts can point whois at a local server instead of the public internet.",
		Examples: []command.Example{
			{Command: "whois example.test", Explain: "Query the default WHOIS server."},
			{Command: "whois -h whois.example.test example.test", Explain: "Query a specific server."},
		},
		ExitStatus: "0  a response was received and printed.\n" +
			"1  bad arguments or the query failed.",
	})
	server := fs.StringP("host", "h", defaultServer, "WHOIS server to query (port 43)")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) != 1 {
		return command.Failuref("exactly one OBJECT (domain or IP) is required")
	}

	resp, err := query(*server, operands[0])
	if err != nil {
		return command.Failuref("%v", err)
	}
	_, _ = fmt.Fprint(stdio.Out, resp)
	return nil
}

// tcpQuery performs the WHOIS protocol: connect to the server (port 43 unless an
// explicit host:port is given), send the object followed by CRLF, and read the
// whole response. Accepting an explicit port lets tests target a loopback stub.
func tcpQuery(server, object string) (string, error) {
	addr := server
	if _, _, err := net.SplitHostPort(server); err != nil {
		addr = net.JoinHostPort(server, "43")
	}
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	if _, err := fmt.Fprintf(conn, "%s\r\n", object); err != nil {
		return "", err
	}

	var b []byte
	sc := bufio.NewScanner(conn)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		b = append(b, sc.Bytes()...)
		b = append(b, '\n')
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return string(b), nil
}
