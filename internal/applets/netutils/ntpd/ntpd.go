// Package ntpd implements the ntpd applet plus the NTP packet codec and
// client-query logic it builds on.
//
// The NTP v4 packet encoder/decoder and the offset calculation are pure
// functions, and the query goes through an injectable Transport so the
// request/response logic can be unit-tested deterministically. Disciplining the
// system clock is privileged and host-mutating, so the host-facing daemon mode
// is capability-gated and fails with a documented error; the query/offset path
// is what the tests exercise.
package ntpd

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// ntpEpochOffset is the difference between the NTP epoch (1900) and the Unix
// epoch (1970) in seconds.
const ntpEpochOffset = 2208988800

// Packet is the 48-byte NTP message this slice cares about.
type Packet struct {
	LeapVersionMode byte
	Stratum         byte
	OriginateTime   uint64 // NTP timestamp
	ReceiveTime     uint64
	TransmitTime    uint64
}

// Marshal encodes the packet into the 48-byte NTP wire format.
func (p *Packet) Marshal() []byte {
	b := make([]byte, 48)
	b[0] = p.LeapVersionMode
	b[1] = p.Stratum
	binary.BigEndian.PutUint64(b[24:32], p.OriginateTime)
	binary.BigEndian.PutUint64(b[32:40], p.ReceiveTime)
	binary.BigEndian.PutUint64(b[40:48], p.TransmitTime)
	return b
}

// errShort marks a truncated NTP packet.
var errShort = errors.New("short NTP packet")

// Unmarshal decodes a 48-byte NTP packet.
func Unmarshal(b []byte) (*Packet, error) {
	if len(b) < 48 {
		return nil, errShort
	}
	return &Packet{
		LeapVersionMode: b[0],
		Stratum:         b[1],
		OriginateTime:   binary.BigEndian.Uint64(b[24:32]),
		ReceiveTime:     binary.BigEndian.Uint64(b[32:40]),
		TransmitTime:    binary.BigEndian.Uint64(b[40:48]),
	}, nil
}

// ToNTP converts a time.Time to an NTP timestamp (seconds in the high 32 bits,
// fractional seconds in the low 32 bits).
func ToNTP(t time.Time) uint64 {
	secs := uint64(t.Unix() + ntpEpochOffset)
	frac := uint64(t.Nanosecond()) << 32 / 1e9
	return secs<<32 | frac
}

// FromNTP converts an NTP timestamp back to a time.Time (UTC).
func FromNTP(ts uint64) time.Time {
	secs := int64(ts>>32) - ntpEpochOffset
	frac := int64((ts & 0xffffffff) * 1e9 >> 32)
	return time.Unix(secs, frac).UTC()
}

// NewClientRequest builds a client-mode (mode 3, version 4) request stamped with
// the given transmit time.
func NewClientRequest(now time.Time) *Packet {
	return &Packet{
		LeapVersionMode: 0x23, // LI=0, VN=4, Mode=3 (client)
		TransmitTime:    ToNTP(now),
	}
}

// Transport sends an NTP request and returns the server's reply.
type Transport interface {
	Query(request []byte) (reply []byte, err error)
}

// Offset runs one NTP exchange over t and returns the clock offset between the
// local clock and the server, using the standard NTP offset formula:
// ((T2 - T1) + (T3 - T4)) / 2. nowFn supplies the local clock so tests can drive
// it deterministically; pass time.Now in production. It is called for T1 (the
// transmit time) and again for T4 (the receive time).
func Offset(nowFn func() time.Time, t Transport) (time.Duration, *Packet, error) {
	t1 := nowFn()
	req := NewClientRequest(t1)
	replyBytes, err := t.Query(req.Marshal())
	if err != nil {
		return 0, nil, err
	}
	t4 := nowFn()
	reply, err := Unmarshal(replyBytes)
	if err != nil {
		return 0, nil, err
	}
	t2 := FromNTP(reply.ReceiveTime)
	t3 := FromNTP(reply.TransmitTime)
	offset := ((t2.Sub(t1)) + (t3.Sub(t4))) / 2
	return offset, reply, nil
}

// ServerReply builds an NTP server reply (mode 4) to req, stamping receive and
// transmit times from recvTime and txTime and echoing the originate timestamp.
func ServerReply(req *Packet, recvTime, txTime time.Time, stratum byte) *Packet {
	return &Packet{
		LeapVersionMode: 0x24, // LI=0, VN=4, Mode=4 (server)
		Stratum:         stratum,
		OriginateTime:   req.TransmitTime,
		ReceiveTime:     ToNTP(recvTime),
		TransmitTime:    ToNTP(txTime),
	}
}

// Command is the ntpd applet.
type Command struct{}

// New returns an ntpd command.
func New() *Command { return &Command{} }

// Name returns the command name.
func (c *Command) Name() string { return "ntpd" }

// Synopsis returns the one-line description shown in the applet list.
func (c *Command) Synopsis() string { return "NTP client/daemon (query implemented; clock set gated)" }

// Run executes ntpd.
func (c *Command) Run(_ context.Context, stdio command.IO, args []string) error {
	fs := command.NewFlagSet(c.Name(), "[-q] -p SERVER", stdio.Err).WithHelp(command.Help{
		Description: "NTP client/daemon. The packet codec, client request, and offset calculation are " +
			"implemented and unit-tested via an injectable transport. -p names a server and -q requests a " +
			"one-shot query. Disciplining the system clock (stepping or slewing time) is a privileged, " +
			"host-mutating operation that is not available in this environment, so this applet validates " +
			"its arguments and fails with a documented capability error rather than silently doing nothing.",
		Examples: []command.Example{
			{Command: "ntpd -q -p pool.ntp.org", Explain: "One-shot query (capability-gated for the clock-set step)."},
		},
		ExitStatus: "0  never in this environment.\n1  always: validated request then a documented backend error.",
		Notes: []string{
			"The offset calculation is exercised by transport-injection tests; setting the clock is capability-gated.",
		},
	})
	server := fs.StringP("peer", "p", "", "NTP server to query")
	_ = fs.BoolP("query", "q", false, "query once then exit")

	proceed, err := fs.Parse(stdio, args)
	if err != nil || !proceed {
		return err
	}
	if *server == "" {
		return command.Failuref("an NTP server is required (-p)")
	}
	return command.Failuref(
		"ntpd: would query %q and discipline the clock, but stepping/slewing the system clock is not "+
			"available in this environment (capability-gated backend)", *server)
}

// UDPTransport is the production Transport over UDP, exported so a future slice
// can wire it up to a real NTP query path.
type UDPTransport struct {
	Addr    string
	Timeout time.Duration
}

// Query sends request to the configured NTP server and returns the reply.
func (u *UDPTransport) Query(request []byte) ([]byte, error) {
	conn, err := net.DialTimeout("udp", u.Addr, u.Timeout)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(u.Timeout))
	if _, err := conn.Write(request); err != nil {
		return nil, err
	}
	buf := make([]byte, 48)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

var _ Transport = (*UDPTransport)(nil)
