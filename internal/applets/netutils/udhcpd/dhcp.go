// Package udhcpd implements the udhcpd and dhcprelay applets together with the
// DHCP wire codec they share.
//
// The DHCP message encoder/decoder and the lease/config parsers are pure so they
// can be table-tested, and packet transmission goes through an injectable
// Transport interface so the request/response logic can be unit-tested without
// touching a real socket.
package udhcpd

import (
	"encoding/binary"
	"errors"
	"net"
)

// DHCP message op codes and the magic cookie.
const (
	OpBootRequest = 1
	OpBootReply   = 2
)

// magicCookie precedes the options field of every DHCP message.
var magicCookie = []byte{99, 130, 83, 99}

// DHCP option codes used by this slice.
const (
	OptSubnetMask    = 1
	OptRouter        = 3
	OptDNSServer     = 6
	OptRequestedIP   = 50
	OptLeaseTime     = 51
	OptMessageType   = 53
	OptServerID      = 54
	OptEnd           = 255
)

// DHCP message types (option 53 values).
const (
	Discover = 1
	Offer    = 2
	Request  = 3
	Decline  = 4
	ACK      = 5
	NAK      = 6
	Release  = 7
	Inform   = 8
)

// Message is a decoded DHCP message. Only the fields this slice needs are kept.
type Message struct {
	Op      byte
	XID     uint32
	CIAddr  net.IP // client IP
	YIAddr  net.IP // your (assigned) IP
	SIAddr  net.IP // next server IP
	GIAddr  net.IP // relay agent IP
	CHAddr  net.HardwareAddr
	Options map[byte][]byte
}

// Type returns the DHCP message type (option 53), or 0 when absent.
func (m *Message) Type() byte {
	if v, ok := m.Options[OptMessageType]; ok && len(v) == 1 {
		return v[0]
	}
	return 0
}

// Marshal encodes m into a DHCP packet.
func (m *Message) Marshal() []byte {
	buf := make([]byte, 240)
	buf[0] = m.Op
	buf[1] = 1  // htype: Ethernet
	buf[2] = 6  // hlen
	binary.BigEndian.PutUint32(buf[4:8], m.XID)
	copyIP(buf[12:16], m.CIAddr)
	copyIP(buf[16:20], m.YIAddr)
	copyIP(buf[20:24], m.SIAddr)
	copyIP(buf[24:28], m.GIAddr)
	if len(m.CHAddr) >= 6 {
		copy(buf[28:34], m.CHAddr[:6])
	}
	copy(buf[236:240], magicCookie)

	// Options, with message type first when present.
	if v, ok := m.Options[OptMessageType]; ok {
		buf = append(buf, OptMessageType, byte(len(v)))
		buf = append(buf, v...)
	}
	for code, val := range m.Options {
		if code == OptMessageType || code == OptEnd {
			continue
		}
		buf = append(buf, code, byte(len(val)))
		buf = append(buf, val...)
	}
	buf = append(buf, OptEnd)
	return buf
}

// errShort marks a truncated DHCP packet.
var errShort = errors.New("short DHCP packet")

// Unmarshal decodes a DHCP packet into a Message.
func Unmarshal(p []byte) (*Message, error) {
	if len(p) < 240 {
		return nil, errShort
	}
	if !equal(p[236:240], magicCookie) {
		return nil, errors.New("bad DHCP magic cookie")
	}
	m := &Message{
		Op:      p[0],
		XID:     binary.BigEndian.Uint32(p[4:8]),
		CIAddr:  net.IP(append([]byte{}, p[12:16]...)),
		YIAddr:  net.IP(append([]byte{}, p[16:20]...)),
		SIAddr:  net.IP(append([]byte{}, p[20:24]...)),
		GIAddr:  net.IP(append([]byte{}, p[24:28]...)),
		CHAddr:  net.HardwareAddr(append([]byte{}, p[28:34]...)),
		Options: map[byte][]byte{},
	}
	i := 240
	for i < len(p) {
		code := p[i]
		i++
		if code == OptEnd {
			break
		}
		if code == 0 {
			continue // pad
		}
		if i >= len(p) {
			return nil, errShort
		}
		l := int(p[i])
		i++
		if i+l > len(p) {
			return nil, errShort
		}
		m.Options[code] = append([]byte{}, p[i:i+l]...)
		i += l
	}
	return m, nil
}

func copyIP(dst []byte, ip net.IP) {
	if v4 := ip.To4(); v4 != nil {
		copy(dst, v4)
	}
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// LeaseTimeOption encodes a lease time (seconds) as option 51 data.
func LeaseTimeOption(seconds uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, seconds)
	return b
}
