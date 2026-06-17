// Package memnet provides in-memory network primitives for hermetic unit tests
// of the netutils applets. It lets tests exercise accept/serve loops and
// client/server protocol handlers without binding real loopback sockets, so the
// vast majority of netutils tests no longer have to skip when sockets are
// unavailable (sandboxes, restricted CI, etc.).
//
// The two main types are:
//
//   - PipeListener: a net.Listener whose Accept returns the server side of an
//     in-memory net.Pipe. Tests obtain client connections via Dial. Use this for
//     applets whose Serve loop takes a net.Listener.
//
//   - PacketPipe: a pair of in-memory net.PacketConn endpoints connected back to
//     back. Use this for UDP-style applets whose serve loop takes a PacketConn
//     interface (net.Pipe is stream-only and cannot model datagrams).
//
// Everything here is intended for use from _test.go files; it lives in a
// non-test package only so it can be shared across applet packages.
package memnet

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

// PipeListener is an in-memory net.Listener. Each Dial is matched with one
// Accept, handing both sides a connected, buffered net.Conn (see BufferedConn).
// Closing the listener unblocks any pending Accept with an error, which mirrors
// how a real listener behaves when closed during shutdown.
type PipeListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

// NewPipeListener returns a ready-to-use in-memory listener.
func NewPipeListener() *PipeListener {
	return &PipeListener{
		conns:  make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

// pipeAddr is a stand-in address for in-memory connections.
type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "memnet" }

// Accept returns the server side of the next dialed connection. It blocks until
// a Dial occurs or the listener is closed.
func (l *PipeListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.closed:
		return nil, errClosed
	}
}

// Close stops the listener. Subsequent Accept calls return an error.
func (l *PipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

// Addr returns the listener's pseudo-address.
func (l *PipeListener) Addr() net.Addr { return pipeAddr{} }

// Dial creates a connected pair and delivers the server side to a pending (or
// future) Accept, returning the client side to the caller. It fails if the
// listener has been closed. The returned connections support CloseWrite (via
// the halfCloser/HalfCloseConn pattern), so server handlers that read to EOF and
// then write a reply behave the same as they would over a real TCP socket.
func (l *PipeListener) Dial() (net.Conn, error) {
	client, server := BufferedConn()
	select {
	case l.conns <- server:
		return client, nil
	case <-l.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, errClosed
	}
}

// HalfCloseConn is a net.Conn that also supports closing only the write half,
// which lets the peer observe EOF on reads while the connection stays open for
// the reply direction. It is satisfied by the connections returned from
// BufferedConn (and thus by PipeListener.Dial / Accept).
type HalfCloseConn interface {
	net.Conn
	CloseWrite() error
}

// errClosed is returned by operations on a closed PipeListener or PacketPipe.
var errClosed = errors.New("memnet: closed")

// PacketPipe is one endpoint of an in-memory, bidirectional packet channel that
// satisfies net.PacketConn. Datagrams written to one endpoint are readable from
// the other, preserving message boundaries (unlike net.Pipe). It is sufficient
// to drive datagram serve loops (DNS, TFTP, NTP, ...) entirely in memory.
type PacketPipe struct {
	addr     packetAddr
	peer     packetAddr
	in       chan datagram
	out      chan datagram
	closed   chan struct{}
	once     sync.Once
	deadline atomicTime
}

// datagram is a single message plus its source address.
type datagram struct {
	data []byte
	from net.Addr
}

// packetAddr is a named in-memory packet address.
type packetAddr string

func (a packetAddr) Network() string { return "mempacket" }
func (a packetAddr) String() string  { return string(a) }

// NewPacketPipe returns two connected in-memory PacketConn endpoints. A write to
// the first is read from the second and vice versa.
func NewPacketPipe() (*PacketPipe, *PacketPipe) {
	a2b := make(chan datagram, 16)
	b2a := make(chan datagram, 16)
	closed := make(chan struct{})
	a := &PacketPipe{addr: "a", peer: "b", in: b2a, out: a2b, closed: closed}
	b := &PacketPipe{addr: "b", peer: "a", in: a2b, out: b2a, closed: closed}
	return a, b
}

// ReadFrom reads one datagram, returning its payload and source address.
func (p *PacketPipe) ReadFrom(b []byte) (int, net.Addr, error) {
	var timeout <-chan time.Time
	if d := p.deadline.Load(); !d.IsZero() {
		timer := time.NewTimer(time.Until(d))
		defer timer.Stop()
		timeout = timer.C
	}
	select {
	case dg, ok := <-p.in:
		if !ok {
			return 0, nil, errClosed
		}
		n := copy(b, dg.data)
		return n, dg.from, nil
	case <-p.closed:
		return 0, nil, errClosed
	case <-timeout:
		return 0, nil, timeoutError{}
	}
}

// WriteTo sends b as one datagram to the peer endpoint. The destination address
// argument is accepted for interface compatibility but ignored: there is exactly
// one peer.
func (p *PacketPipe) WriteTo(b []byte, _ net.Addr) (int, error) {
	dg := datagram{data: append([]byte(nil), b...), from: p.addr}
	select {
	case p.out <- dg:
		return len(b), nil
	case <-p.closed:
		return 0, errClosed
	}
}

// Close shuts down both endpoints.
func (p *PacketPipe) Close() error {
	p.once.Do(func() { close(p.closed) })
	return nil
}

// LocalAddr returns this endpoint's pseudo-address.
func (p *PacketPipe) LocalAddr() net.Addr { return p.addr }

// PeerAddr returns the address of the connected endpoint; useful as the "to"
// argument for WriteTo when symmetry is desired.
func (p *PacketPipe) PeerAddr() net.Addr { return p.peer }

// SetDeadline sets both read and write deadlines (only reads honor it here).
func (p *PacketPipe) SetDeadline(t time.Time) error {
	p.deadline.Store(t)
	return nil
}

// SetReadDeadline sets the read deadline.
func (p *PacketPipe) SetReadDeadline(t time.Time) error {
	p.deadline.Store(t)
	return nil
}

// SetWriteDeadline is a no-op (writes never block past Close here).
func (p *PacketPipe) SetWriteDeadline(time.Time) error { return nil }

// BufferedConn returns a connected pair of in-memory, full-duplex connections
// with internal buffering in each direction. Unlike net.Pipe (which is fully
// synchronous and can deadlock protocols that write interleaved records before
// reading, such as a TLS handshake), each Write here returns as soon as the
// bytes are buffered. Each side can close its write half via CloseWrite.
func BufferedConn() (HalfCloseConn, HalfCloseConn) {
	a2b := newByteChan()
	b2a := newByteChan()
	a := &bufConn{rd: b2a, wr: a2b}
	b := &bufConn{rd: a2b, wr: b2a}
	return a, b
}

// byteChan is an unbounded, closable in-memory byte stream protected by a mutex
// and signalled by a condition variable.
type byteChan struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	closed bool
}

func newByteChan() *byteChan {
	bc := &byteChan{}
	bc.cond = sync.NewCond(&bc.mu)
	return bc
}

func (bc *byteChan) write(p []byte) (int, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if bc.closed {
		return 0, io.ErrClosedPipe
	}
	bc.buf = append(bc.buf, p...)
	bc.cond.Broadcast()
	return len(p), nil
}

func (bc *byteChan) read(p []byte) (int, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	for len(bc.buf) == 0 {
		if bc.closed {
			return 0, io.EOF
		}
		bc.cond.Wait()
	}
	n := copy(p, bc.buf)
	bc.buf = bc.buf[n:]
	return n, nil
}

func (bc *byteChan) close() {
	bc.mu.Lock()
	bc.closed = true
	bc.cond.Broadcast()
	bc.mu.Unlock()
}

// bufConn is one end of a BufferedConn pair.
type bufConn struct {
	rd *byteChan
	wr *byteChan
}

func (c *bufConn) Read(p []byte) (int, error)  { return c.rd.read(p) }
func (c *bufConn) Write(p []byte) (int, error) { return c.wr.write(p) }

func (c *bufConn) Close() error {
	c.wr.close()
	c.rd.close()
	return nil
}

func (c *bufConn) CloseWrite() error { c.wr.close(); return nil }

func (c *bufConn) LocalAddr() net.Addr         { return pipeAddr{} }
func (c *bufConn) RemoteAddr() net.Addr        { return pipeAddr{} }
func (c *bufConn) SetDeadline(time.Time) error { return nil }
func (c *bufConn) SetReadDeadline(time.Time) error {
	return nil
}
func (c *bufConn) SetWriteDeadline(time.Time) error { return nil }

// timeoutError satisfies net.Error with Timeout() == true.
type timeoutError struct{}

func (timeoutError) Error() string   { return "memnet: i/o timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

// atomicTime is a tiny mutex-guarded time.Time for deadline storage.
type atomicTime struct {
	mu sync.Mutex
	t  time.Time
}

func (a *atomicTime) Store(t time.Time) {
	a.mu.Lock()
	a.t = t
	a.mu.Unlock()
}

func (a *atomicTime) Load() time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.t
}
