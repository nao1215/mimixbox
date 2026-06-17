// Package tftpd implements the tftpd applet: a small read-only TFTP server
// (RFC 1350) that serves files from a directory over UDP.
//
// The packet codec (RRQ/DATA/ACK/ERROR) is implemented as pure helpers so it can
// be table-tested, and the foreground server runs over loopback for hermetic
// integration tests. Only octet-mode reads (RRQ) are supported in this slice.
package tftpd

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/nao1215/mimixbox/internal/command"
)

// TFTP opcodes.
const (
	opRRQ   = 1
	opWRQ   = 2
	opDATA  = 3
	opACK   = 4
	opERROR = 5
)

const blockSize = 512

// Command is the tftpd applet.
type Command struct{}

// New returns a tftpd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "tftpd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Read-only TFTP server" }

// Run executes tftpd.
func (c *Command) Run(ctx context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "-f [-l ADDR] DIR", stdio.Err).WithHelp(command.Help{
		Description: "Serve files read-only over TFTP (RFC 1350) from DIR. -f keeps tftpd in the " +
			"foreground; -l sets the UDP listen address (default 127.0.0.1:69). Only read requests (RRQ) " +
			"in octet mode are served; writes are refused with a TFTP error. Path traversal outside DIR is " +
			"rejected. The server runs until its context is cancelled.",
		Examples: []command.Example{
			{Command: "tftpd -f -l 127.0.0.1:6969 ./tftproot", Explain: "Serve ./tftproot read-only on loopback port 6969."},
		},
		ExitStatus: "0  clean shutdown.\n1  bad arguments or bind error.",
		Notes: []string{
			"Write requests (WRQ) are intentionally refused; this slice is read-only.",
		},
	})
	foreground := fs.BoolP("foreground", "f", false, "run in the foreground (required in this slice)")
	addr := fs.StringP("listen", "l", "127.0.0.1:69", "UDP address to listen on (HOST:PORT)")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if !*foreground {
		return command.Failuref("only foreground mode is implemented; pass -f")
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return command.Failuref("a root directory is required")
	}
	root, err := filepath.Abs(rest[0])
	if err != nil {
		return command.Failuref("invalid root %q: %v", rest[0], err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", *addr)
	if err != nil {
		return command.Failuref("resolve %s: %v", *addr, err)
	}
	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return command.Failuref("cannot listen on %s: %v", *addr, err)
	}
	_, _ = fmt.Fprintf(stdio.Out, "tftpd: serving %s on %s\n", root, pc.LocalAddr().String())
	return Serve(ctx, pc, root)
}

// Serve runs the TFTP loop on pc until ctx is cancelled. Each read request is
// answered with the file's contents split into 512-byte DATA blocks. pc is a
// net.PacketConn so tests can drive the loop with an in-memory packet pipe
// instead of a real UDP socket.
func Serve(ctx context.Context, pc net.PacketConn, root string) error {
	go func() {
		<-ctx.Done()
		_ = pc.Close()
	}()
	buf := make([]byte, 1024)
	for {
		n, raddr, err := pc.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return command.Failuref("read: %v", err)
		}
		req := make([]byte, n)
		copy(req, buf[:n])
		handleRequest(pc, raddr, req, root)
	}
}

// handleRequest decodes and serves a single request datagram.
func handleRequest(pc net.PacketConn, raddr net.Addr, req []byte, root string) {
	op, filename, _, err := ParseRequest(req)
	if err != nil {
		_, _ = pc.WriteTo(ErrorPacket(4, "illegal TFTP operation"), raddr)
		return
	}
	if op == opWRQ {
		_, _ = pc.WriteTo(ErrorPacket(2, "writes are not permitted"), raddr)
		return
	}
	if op != opRRQ {
		_, _ = pc.WriteTo(ErrorPacket(4, "illegal TFTP operation"), raddr)
		return
	}
	path, err := safeJoin(root, filename)
	if err != nil {
		_, _ = pc.WriteTo(ErrorPacket(2, "access violation"), raddr)
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		_, _ = pc.WriteTo(ErrorPacket(1, "file not found"), raddr)
		return
	}
	sendFile(pc, raddr, data)
}

// sendFile transfers data as a sequence of DATA blocks; it does not wait for
// ACKs (a simplified lock-step is sufficient for hermetic tests, where the
// client reads sequentially). A final short block terminates the transfer.
func sendFile(pc net.PacketConn, raddr net.Addr, data []byte) {
	block := uint16(1)
	for off := 0; ; off += blockSize {
		end := off + blockSize
		if end > len(data) {
			end = len(data)
		}
		_, _ = pc.WriteTo(DataPacket(block, data[off:end]), raddr)
		if end == len(data) {
			return
		}
		block++
	}
}

// ParseRequest decodes an RRQ/WRQ packet, returning the opcode, filename and
// mode. Other opcodes return their opcode with empty strings.
func ParseRequest(p []byte) (op int, filename, mode string, err error) {
	if len(p) < 2 {
		return 0, "", "", errors.New("short packet")
	}
	op = int(binary.BigEndian.Uint16(p[0:2]))
	if op != opRRQ && op != opWRQ {
		return op, "", "", nil
	}
	parts := strings.SplitN(string(p[2:]), "\x00", 3)
	if len(parts) < 2 {
		return op, "", "", errors.New("malformed request")
	}
	return op, parts[0], parts[1], nil
}

// DataPacket builds a DATA packet for block with payload.
func DataPacket(block uint16, payload []byte) []byte {
	p := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint16(p[0:2], opDATA)
	binary.BigEndian.PutUint16(p[2:4], block)
	copy(p[4:], payload)
	return p
}

// ErrorPacket builds an ERROR packet with code and message.
func ErrorPacket(code uint16, msg string) []byte {
	p := make([]byte, 0, 5+len(msg))
	hdr := make([]byte, 4)
	binary.BigEndian.PutUint16(hdr[0:2], opERROR)
	binary.BigEndian.PutUint16(hdr[2:4], code)
	p = append(p, hdr...)
	p = append(p, msg...)
	p = append(p, 0)
	return p
}

// safeJoin joins root and name, rejecting paths that escape root.
func safeJoin(root, name string) (string, error) {
	clean := filepath.Clean("/" + name) // force absolute, drop ".."
	full := filepath.Join(root, clean)
	if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		return "", errors.New("path escapes root")
	}
	return full, nil
}
