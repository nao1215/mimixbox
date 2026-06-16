package tcpsvd

import (
	"io"
	"net"
	"time"
)

// udpConn adapts a single received datagram to the net.Conn interface so the
// same ConnHandler works for both TCP and UDP. Read yields the datagram payload
// once; Write sends a reply datagram back to the original sender.
type udpConn struct {
	pc      *net.UDPConn
	raddr   *net.UDPAddr
	payload []byte
	off     int
}

func newUDPConn(pc *net.UDPConn, raddr *net.UDPAddr, payload []byte) *udpConn {
	return &udpConn{pc: pc, raddr: raddr, payload: payload}
}

func (u *udpConn) Read(p []byte) (int, error) {
	if u.off >= len(u.payload) {
		return 0, io.EOF
	}
	n := copy(p, u.payload[u.off:])
	u.off += n
	return n, nil
}

func (u *udpConn) Write(p []byte) (int, error)       { return u.pc.WriteToUDP(p, u.raddr) }
func (u *udpConn) Close() error                      { return nil }
func (u *udpConn) LocalAddr() net.Addr               { return u.pc.LocalAddr() }
func (u *udpConn) RemoteAddr() net.Addr              { return u.raddr }
func (u *udpConn) SetDeadline(_ time.Time) error     { return nil }
func (u *udpConn) SetReadDeadline(_ time.Time) error { return nil }
func (u *udpConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

var _ net.Conn = (*udpConn)(nil)
