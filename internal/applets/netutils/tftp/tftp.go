// Package tftp implements the tftp applet: a minimal RFC 1350 TFTP client that
// can get (download) or put (upload) a file in octet mode over UDP. The
// transport is exercised against a loopback fixture server in tests, so no
// external network is needed. Only octet mode and 512-byte blocks are supported;
// options (blksize, tsize, timeout negotiation) are intentionally not
// implemented.
package tftp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// TFTP opcodes (RFC 1350).
const (
	opRRQ   = 1
	opWRQ   = 2
	opDATA  = 3
	opACK   = 4
	opERROR = 5
)

const blockSize = 512

// Command is the tftp applet.
type Command struct{}

// New returns a tftp command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tftp" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Transfer a file with a TFTP server (get/put)" }

// openFile / createFile are injectable so tests can use in-memory buffers
// instead of touching the filesystem.
var (
	readLocal  = os.ReadFile
	writeLocal = os.WriteFile
)

// Run executes tftp.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-g|-p] -l LOCAL -r REMOTE HOST [PORT]", stdio.Err).WithHelp(command.Help{
		Description: "Transfer a single file to or from a TFTP server over UDP in octet (binary) mode. " +
			"Use -g to get (download) the REMOTE file into LOCAL, or -p to put (upload) the LOCAL " +
			"file as REMOTE. HOST is the server; PORT defaults to 69. Only 512-byte block octet mode " +
			"is implemented; TFTP option negotiation (blksize/tsize/timeout) is not.",
		Examples: []command.Example{
			{Command: "tftp -g -l out.bin -r firmware.bin 127.0.0.1", Explain: "Download firmware.bin."},
			{Command: "tftp -p -l config.txt -r config.txt 127.0.0.1 6900", Explain: "Upload to a custom port."},
		},
		ExitStatus: "0  the transfer completed.\n" +
			"1  bad arguments or a transfer error.",
		Notes: []string{"Only octet mode with 512-byte blocks is supported; options are not negotiated."},
	})
	get := fs.BoolP("get", "g", false, "download REMOTE into LOCAL")
	put := fs.BoolP("put", "p", false, "upload LOCAL as REMOTE")
	local := fs.StringP("local", "l", "", "local file path")
	remote := fs.StringP("remote", "r", "", "remote file name")
	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	if *get == *put {
		return command.Failuref("exactly one of -g (get) or -p (put) is required")
	}
	if *local == "" || *remote == "" {
		return command.Failuref("both -l LOCAL and -r REMOTE are required")
	}

	host, port, err := hostPort(fs.Args())
	if err != nil {
		return command.Failuref("%v", err)
	}
	server := net.JoinHostPort(host, port)

	if *get {
		data, err := tftpGet(server, *remote)
		if err != nil {
			return command.Failuref("get failed: %v", err)
		}
		if err := writeLocal(*local, data, 0o644); err != nil {
			return command.Failuref("cannot write %q: %v", *local, err)
		}
		fmt.Fprintf(stdio.Out, "Received %d bytes into %s\n", len(data), *local)
		return nil
	}

	data, err := readLocal(*local)
	if err != nil {
		return command.Failuref("cannot read %q: %v", *local, err)
	}
	if err := tftpPut(server, *remote, data); err != nil {
		return command.Failuref("put failed: %v", err)
	}
	fmt.Fprintf(stdio.Out, "Sent %d bytes from %s\n", len(data), *local)
	return nil
}

// hostPort extracts HOST and PORT (default 69) from the operands.
func hostPort(operands []string) (host, port string, err error) {
	switch len(operands) {
	case 1:
		return operands[0], "69", nil
	case 2:
		return operands[0], operands[1], nil
	default:
		return "", "", fmt.Errorf("usage: tftp [-g|-p] -l LOCAL -r REMOTE HOST [PORT]")
	}
}

// dialUDP connects to server; split out so the deadline handling is shared.
func dialUDP(server string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	return conn, nil
}

// tftpGet downloads remote via an RRQ and returns the assembled file.
func tftpGet(server, remote string) ([]byte, error) {
	conn, err := dialUDP(server)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write(request(opRRQ, remote)); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	expected := uint16(1)
	buf := make([]byte, 4+blockSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		op, block, payload, errMsg := parsePacket(buf[:n])
		if op == opERROR {
			return nil, fmt.Errorf("server error: %s", errMsg)
		}
		if op != opDATA {
			return nil, fmt.Errorf("unexpected opcode %d", op)
		}
		if block == expected {
			out.Write(payload)
			if _, err := conn.Write(ack(block)); err != nil {
				return nil, err
			}
			expected++
			if len(payload) < blockSize {
				return out.Bytes(), nil
			}
		}
	}
}

// tftpPut uploads data as remote via a WRQ.
func tftpPut(server, remote string, data []byte) error {
	conn, err := dialUDP(server)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write(request(opWRQ, remote)); err != nil {
		return err
	}
	// Expect ACK 0.
	if err := expectAck(conn, 0); err != nil {
		return err
	}

	block := uint16(1)
	for off := 0; ; off += blockSize {
		end := off + blockSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[off:end]
		if _, err := conn.Write(dataPacket(block, chunk)); err != nil {
			return err
		}
		if err := expectAck(conn, block); err != nil {
			return err
		}
		block++
		if len(chunk) < blockSize {
			return nil
		}
	}
}

// expectAck reads one packet and verifies it is an ACK for want.
func expectAck(conn *net.UDPConn, want uint16) error {
	buf := make([]byte, 4+blockSize)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	op, block, _, errMsg := parsePacket(buf[:n])
	if op == opERROR {
		return fmt.Errorf("server error: %s", errMsg)
	}
	if op != opACK || block != want {
		return fmt.Errorf("expected ACK %d, got opcode %d block %d", want, op, block)
	}
	return nil
}

// request builds an RRQ/WRQ packet for filename in octet mode.
func request(op int, filename string) []byte {
	var b bytes.Buffer
	_ = binary.Write(&b, binary.BigEndian, uint16(op))
	b.WriteString(filename)
	b.WriteByte(0)
	b.WriteString("octet")
	b.WriteByte(0)
	return b.Bytes()
}

// ack builds an ACK packet for block.
func ack(block uint16) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:2], opACK)
	binary.BigEndian.PutUint16(b[2:4], block)
	return b
}

// dataPacket builds a DATA packet for block carrying payload.
func dataPacket(block uint16, payload []byte) []byte {
	b := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint16(b[0:2], opDATA)
	binary.BigEndian.PutUint16(b[2:4], block)
	copy(b[4:], payload)
	return b
}

// parsePacket decodes the opcode and the fields relevant to each packet type.
func parsePacket(p []byte) (op int, block uint16, payload []byte, errMsg string) {
	if len(p) < 2 {
		return 0, 0, nil, ""
	}
	op = int(binary.BigEndian.Uint16(p[0:2]))
	switch op {
	case opDATA:
		if len(p) >= 4 {
			block = binary.BigEndian.Uint16(p[2:4])
			payload = p[4:]
		}
	case opACK:
		if len(p) >= 4 {
			block = binary.BigEndian.Uint16(p[2:4])
		}
	case opERROR:
		if len(p) >= 4 {
			msg := p[4:]
			if i := bytes.IndexByte(msg, 0); i >= 0 {
				msg = msg[:i]
			}
			errMsg = string(msg)
		}
	}
	return op, block, payload, errMsg
}
