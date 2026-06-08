// Package ping implements the ping applet: send ICMP ECHO_REQUEST packets to a
// network host and report the round-trip times.
//
// This is a clean-room reimplementation. The original morrigan ping was forked
// from the u-root project (BSD-3-Clause, Copyright the u-root Authors); that
// attribution is preserved here per the BSD-3 terms even though no source is
// copied verbatim.
//
// Sending ICMP echo requests needs a raw socket (CAP_NET_RAW / root). When the
// socket cannot be created the command reports the privilege requirement and
// exits with an error rather than crashing.
package ping

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// Command is the ping applet.
type Command struct{}

// New returns a ping command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ping" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "Send ICMP ECHO_REQUEST to network hosts" }

const (
	icmpEchoRequest = 8
	icmpEchoReply   = 0
)

// Run executes ping.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[OPTION]... HOST", stdio.Err)
	count := fs.IntP("count", "c", 4, "stop after sending COUNT packets")
	interval := fs.Float64P("interval", "i", 1.0, "seconds to wait between packets")
	timeout := fs.Float64P("timeout", "W", 1.0, "seconds to wait for each reply")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}

	hosts := fs.Args()
	if len(hosts) != 1 {
		return command.Failuref("exactly one HOST is required")
	}

	addr, err := net.ResolveIPAddr("ip4", hosts[0])
	if err != nil {
		return command.Failuref("cannot resolve %q: %v", hosts[0], err)
	}

	return c.ping(stdio, addr, *count, time.Duration(*interval*float64(time.Second)), time.Duration(*timeout*float64(time.Second)))
}

// ping sends count echo requests to addr, printing each round-trip time.
func (c *Command) ping(stdio command.IO, addr *net.IPAddr, count int, interval, timeout time.Duration) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_ICMP)
	if err != nil {
		return command.Failuref("cannot open raw ICMP socket (needs CAP_NET_RAW/root): %v", err)
	}
	defer func() { _ = unix.Close(fd) }()

	id := os.Getpid() & 0xffff
	var dst [4]byte
	copy(dst[:], addr.IP.To4())
	sa := &unix.SockaddrInet4{Addr: dst}

	_, _ = fmt.Fprintf(stdio.Out, "PING %s (%s)\n", addr.IP, addr.IP)
	received := 0
	for seq := 0; seq < count; seq++ {
		pkt := echoRequest(id, seq)
		start := time.Now()
		if err := unix.Sendto(fd, pkt, 0, sa); err != nil {
			_, _ = fmt.Fprintf(stdio.Err, "ping: send failed: %v\n", err)
			continue
		}
		if rtt, ok := awaitReply(fd, id, seq, timeout, start); ok {
			received++
			_, _ = fmt.Fprintf(stdio.Out, "%d bytes from %s: icmp_seq=%d time=%.2f ms\n",
				len(pkt), addr.IP, seq, float64(rtt.Microseconds())/1000)
		} else {
			_, _ = fmt.Fprintf(stdio.Out, "request timeout for icmp_seq=%d\n", seq)
		}
		if seq < count-1 {
			time.Sleep(interval)
		}
	}

	_, _ = fmt.Fprintf(stdio.Out, "--- %s ping statistics ---\n%d packets transmitted, %d received\n",
		addr.IP, count, received)
	if received == 0 {
		return &command.ExitError{Code: command.ExitFailure}
	}
	return nil
}

// awaitReply waits up to timeout for an echo reply matching id and seq.
func awaitReply(fd, id, seq int, timeout time.Duration, start time.Time) (time.Duration, bool) {
	tv := unix.NsecToTimeval(int64(timeout))
	_ = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv)

	buf := make([]byte, 1500)
	deadline := start.Add(timeout)
	for time.Now().Before(deadline) {
		n, _, err := unix.Recvfrom(fd, buf, 0)
		if err != nil {
			return 0, false
		}
		// Skip the IPv4 header to reach the ICMP message.
		if rid, rseq, ok := parseReply(buf[:n]); ok && rid == id && rseq == seq {
			return time.Since(start), true
		}
	}
	return 0, false
}

// echoRequest builds an ICMP echo request packet with the given id and seq.
func echoRequest(id, seq int) []byte {
	body := make([]byte, 16) // payload
	pkt := make([]byte, 8+len(body))
	pkt[0] = icmpEchoRequest
	pkt[1] = 0
	binary.BigEndian.PutUint16(pkt[4:], uint16(id))
	binary.BigEndian.PutUint16(pkt[6:], uint16(seq))
	copy(pkt[8:], body)
	cs := checksum(pkt)
	binary.BigEndian.PutUint16(pkt[2:], cs)
	return pkt
}

// parseReply extracts the id and seq from a received IPv4+ICMP echo reply.
func parseReply(packet []byte) (id, seq int, ok bool) {
	if len(packet) < 20 {
		return 0, 0, false
	}
	ihl := int(packet[0]&0x0f) * 4
	if len(packet) < ihl+8 {
		return 0, 0, false
	}
	icmp := packet[ihl:]
	if icmp[0] != icmpEchoReply {
		return 0, 0, false
	}
	id = int(binary.BigEndian.Uint16(icmp[4:]))
	seq = int(binary.BigEndian.Uint16(icmp[6:]))
	return id, seq, true
}

// checksum computes the 16-bit one's-complement Internet checksum of data.
func checksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}
