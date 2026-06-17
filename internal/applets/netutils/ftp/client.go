// Package ftp implements the ftpget and ftpput applets: a minimal RFC 959 FTP
// client that downloads (RETR) or uploads (STOR) one file in binary mode over a
// passive (PASV) data connection. Both applets share the same client core and
// differ only in transfer direction. The client is exercised against a loopback
// fixture FTP server in tests, so no external network is needed. Active mode and
// TLS are intentionally not implemented.
//
// This file holds the shared transport/session layer (control-connection setup,
// login, mode selection, and the PASV data-channel transfers) used by both the
// ftpget and ftpput CLI surfaces. Each CLI lives in its own file (ftpget.go,
// ftpput.go) and differs only in transfer direction.
package ftp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// direction selects whether the applet gets or puts a file.
type direction int

const (
	dirGet direction = iota
	dirPut
)

// Command is one FTP applet (ftpget or ftpput).
type Command struct {
	name string
	dir  direction
}

// Name returns the command name.
func (c *Command) Name() string { return c.name }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string {
	if c.dir == dirGet {
		return "Download a file from an FTP server"
	}
	return "Upload a file to an FTP server"
}

// readLocal / writeLocal are injectable so tests avoid the filesystem.
var (
	readLocal  = os.ReadFile
	writeLocal = os.WriteFile
)

// dialTCP opens a TCP connection to address. It is a package-level seam used for
// both the control connection and each PASV data connection, so tests can serve
// FTP over in-memory pipes (returning the right pipe end for the control versus
// data address) without binding loopback sockets. Production dials TCP.
var dialTCP = func(address string) (net.Conn, error) { return net.DialTimeout("tcp", address, 5*time.Second) }

// Run parses the shared CLI surface and drives a transfer in the applet's
// direction. ftpget and ftpput share this entry point and the transport below;
// only the verb wording and the get/put branch differ.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	verb := "Download"
	if c.dir == dirPut {
		verb = "Upload"
	}
	fs := command.NewFlagSet(c.Name(), "[-u USER] [-p PASS] [-P PORT] HOST LOCAL [REMOTE]", stdio.Err).
		WithHelp(command.Help{
			Description: verb + " a single file over FTP in binary mode using a passive (PASV) data " +
				"connection. HOST is the server; LOCAL is the local file path; REMOTE is the remote " +
				"file name (defaults to LOCAL's value). Credentials default to anonymous; override them " +
				"with -u/-p. The control port defaults to 21 (override with -P). Active mode and FTPS " +
				"are not implemented.",
			Examples: []command.Example{
				{Command: c.Name() + " ftp.example.test file.bin", Explain: verb + " file.bin anonymously."},
				{Command: c.Name() + " -u bob -p secret ftp.example.test a.txt b.txt",
					Explain: "Authenticate and use a different remote name."},
			},
			ExitStatus: "0  the transfer completed.\n" +
				"1  bad arguments or a transfer error.",
		})
	user := fs.StringP("username", "u", "anonymous", "FTP username")
	pass := fs.StringP("password", "p", "anonymous@", "FTP password")
	port := fs.StringP("port", "P", "21", "FTP control port")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	operands := fs.Args()
	if len(operands) < 2 || len(operands) > 3 {
		return command.Failuref("usage: %s HOST LOCAL [REMOTE]", c.Name())
	}
	host, local := operands[0], operands[1]
	remote := local
	if len(operands) == 3 {
		remote = operands[2]
	}

	cl, err := dialControl(net.JoinHostPort(host, *port))
	if err != nil {
		return command.Failuref("cannot connect to %s: %v", host, err)
	}
	defer cl.close()

	if err := cl.login(*user, *pass); err != nil {
		return command.Failuref("login failed: %v", err)
	}
	if err := cl.binary(); err != nil {
		return command.Failuref("%v", err)
	}

	if c.dir == dirGet {
		data, err := cl.retrieve(remote)
		if err != nil {
			return command.Failuref("download failed: %v", err)
		}
		if err := writeLocal(local, data, 0o644); err != nil {
			return command.Failuref("cannot write %q: %v", local, err)
		}
		fmt.Fprintf(stdio.Out, "Downloaded %d bytes to %s\n", len(data), local)
		return nil
	}

	data, err := readLocal(local)
	if err != nil {
		return command.Failuref("cannot read %q: %v", local, err)
	}
	if err := cl.store(remote, data); err != nil {
		return command.Failuref("upload failed: %v", err)
	}
	fmt.Fprintf(stdio.Out, "Uploaded %d bytes from %s\n", len(data), local)
	return nil
}

// client wraps an FTP control connection.
type client struct {
	conn net.Conn
	r    *bufio.Reader
}

// dialControl connects to the FTP control port and reads the greeting.
func dialControl(address string) (*client, error) {
	conn, err := dialTCP(address)
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	c := &client{conn: conn, r: bufio.NewReader(conn)}
	if _, _, err := c.readResponse(); err != nil { // greeting
		_ = conn.Close()
		return nil, err
	}
	return c, nil
}

func (c *client) close() { _ = c.conn.Close() }

// cmd sends a command line and returns the response code and text.
func (c *client) cmd(format string, a ...any) (int, string, error) {
	if _, err := fmt.Fprintf(c.conn, format+"\r\n", a...); err != nil {
		return 0, "", err
	}
	return c.readResponse()
}

// readResponse reads a (possibly multi-line) FTP response and returns its code.
func (c *client) readResponse() (int, string, error) {
	line, err := c.r.ReadString('\n')
	if err != nil {
		return 0, "", err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) < 4 {
		return 0, line, fmt.Errorf("short response %q", line)
	}
	code, err := strconv.Atoi(line[:3])
	if err != nil {
		return 0, line, fmt.Errorf("bad response %q", line)
	}
	// Multi-line responses use "code-" on the first line and "code " to end.
	if line[3] == '-' {
		for {
			next, err := c.r.ReadString('\n')
			if err != nil {
				return 0, line, err
			}
			next = strings.TrimRight(next, "\r\n")
			if len(next) >= 4 && next[:3] == line[:3] && next[3] == ' ' {
				break
			}
		}
	}
	return code, line[4:], nil
}

// login performs USER/PASS.
func (c *client) login(user, pass string) error {
	code, _, err := c.cmd("USER %s", user)
	if err != nil {
		return err
	}
	if code == 331 { // need password
		code, _, err = c.cmd("PASS %s", pass)
		if err != nil {
			return err
		}
	}
	if code != 230 {
		return fmt.Errorf("unexpected login response %d", code)
	}
	return nil
}

// binary switches to TYPE I (image/binary) mode.
func (c *client) binary() error {
	code, _, err := c.cmd("TYPE I")
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("cannot set binary mode (response %d)", code)
	}
	return nil
}

// pasv enters passive mode and returns the data-connection address.
func (c *client) pasv() (string, error) {
	code, msg, err := c.cmd("PASV")
	if err != nil {
		return "", err
	}
	if code != 227 {
		return "", fmt.Errorf("PASV failed (response %d)", code)
	}
	return parsePasv(msg)
}

// retrieve downloads remote via RETR over a PASV data connection.
func (c *client) retrieve(remote string) ([]byte, error) {
	dataAddr, err := c.pasv()
	if err != nil {
		return nil, err
	}
	dconn, err := dialTCP(dataAddr)
	if err != nil {
		return nil, err
	}
	defer func() { _ = dconn.Close() }()

	code, _, err := c.cmd("RETR %s", remote)
	if err != nil {
		return nil, err
	}
	if code != 150 && code != 125 {
		return nil, fmt.Errorf("RETR refused (response %d)", code)
	}
	data, err := io.ReadAll(dconn)
	if err != nil {
		return nil, err
	}
	_ = dconn.Close()
	if _, _, err := c.readResponse(); err != nil { // transfer complete
		return nil, err
	}
	return data, nil
}

// store uploads data as remote via STOR over a PASV data connection.
func (c *client) store(remote string, data []byte) error {
	dataAddr, err := c.pasv()
	if err != nil {
		return err
	}
	dconn, err := dialTCP(dataAddr)
	if err != nil {
		return err
	}

	code, _, err := c.cmd("STOR %s", remote)
	if err != nil {
		_ = dconn.Close()
		return err
	}
	if code != 150 && code != 125 {
		_ = dconn.Close()
		return fmt.Errorf("STOR refused (response %d)", code)
	}
	if _, err := dconn.Write(data); err != nil {
		_ = dconn.Close()
		return err
	}
	_ = dconn.Close()
	if _, _, err := c.readResponse(); err != nil { // transfer complete
		return err
	}
	return nil
}

// parsePasv extracts "h1,h2,h3,h4,p1,p2" from a 227 response and builds the
// host:port data-connection address.
func parsePasv(msg string) (string, error) {
	open := strings.IndexByte(msg, '(')
	closeP := strings.IndexByte(msg, ')')
	if open < 0 || closeP < 0 || closeP < open {
		return "", fmt.Errorf("malformed PASV response %q", msg)
	}
	parts := strings.Split(msg[open+1:closeP], ",")
	if len(parts) != 6 {
		return "", fmt.Errorf("malformed PASV tuple %q", msg)
	}
	nums := make([]int, 6)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 || n > 255 {
			return "", fmt.Errorf("bad PASV octet %q", p)
		}
		nums[i] = n
	}
	host := fmt.Sprintf("%d.%d.%d.%d", nums[0], nums[1], nums[2], nums[3])
	port := nums[4]*256 + nums[5]
	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}
