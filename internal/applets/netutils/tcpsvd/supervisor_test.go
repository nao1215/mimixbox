package tcpsvd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeConn is a minimal net.Conn whose Close is observable, used to exercise the
// supervisor's ownership-based connection management without real sockets.
type fakeConn struct {
	closed bool
	addr   string
}

func (f *fakeConn) Read(p []byte) (int, error)         { return 0, io.EOF }
func (f *fakeConn) Write(p []byte) (int, error)        { return len(p), nil }
func (f *fakeConn) Close() error                       { f.closed = true; return nil }
func (f *fakeConn) LocalAddr() net.Addr                { return fakeAddr(f.addr) }
func (f *fakeConn) RemoteAddr() net.Addr               { return fakeAddr(f.addr) }
func (f *fakeConn) SetDeadline(_ time.Time) error      { return nil }
func (f *fakeConn) SetReadDeadline(_ time.Time) error  { return nil }
func (f *fakeConn) SetWriteDeadline(_ time.Time) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return "fake" }
func (a fakeAddr) String() string  { return string(a) }

// fakeCloser records whether the supervisor closed the socket on cancellation.
type fakeCloser struct {
	mu     sync.Mutex
	closed bool
}

func (f *fakeCloser) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeCloser) wasClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

// TestSupervisorServe drives the shared accept loop with an injected accept
// source: it delivers one owned and one unowned connection (verifying that only
// the owned connection is closed by the supervisor and that the handler runs in
// both cases), then signals the loop to drain via a context-cancellation error.
func TestSupervisorServe(t *testing.T) {
	t.Parallel()

	owned := &fakeConn{addr: "owned"}
	unowned := &fakeConn{addr: "unowned"}
	ctx, cancel := context.WithCancel(context.Background())

	var step int
	var handled []string
	var hmu sync.Mutex
	handler := func(c net.Conn) error {
		hmu.Lock()
		handled = append(handled, c.RemoteAddr().String())
		hmu.Unlock()
		return nil
	}

	sock := &fakeCloser{}
	accept := func() (net.Conn, bool, error) {
		switch step {
		case 0:
			step++
			return owned, true, nil
		case 1:
			step++
			return unowned, false, nil
		default:
			// After delivering both connections, cancel and report the error
			// the real loops see once the socket is closed.
			cancel()
			return nil, false, errors.New("use of closed network connection")
		}
	}

	stderr := &bytes.Buffer{}
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: stderr}

	s := supervisor{
		sock:     sock,
		verbose:  true,
		logLine:  func(c net.Conn) string { return "svd: from " + c.RemoteAddr().String() },
		errLabel: "accept",
		accept:   accept,
		handler:  handler,
	}
	if err := s.serve(ctx, stdio); err != nil {
		t.Fatalf("serve returned error after cancel: %v", err)
	}

	// The owned (stream) connection must be closed by the supervisor; the
	// unowned (datagram) connection must not.
	if !owned.closed {
		t.Error("owned connection was not closed by supervisor")
	}
	if unowned.closed {
		t.Error("unowned connection should not be closed by supervisor")
	}

	hmu.Lock()
	defer hmu.Unlock()
	if len(handled) != 2 {
		t.Fatalf("handler ran %d times, want 2 (handled=%v)", len(handled), handled)
	}
	// The ctx.Done goroutine closes the socket asynchronously; in real use that
	// close is what unblocks accept, so poll briefly for it here.
	closed := false
	for i := 0; i < 100; i++ {
		if sock.wasClosed() {
			closed = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !closed {
		t.Error("socket was not closed on context cancellation")
	}
	if stderr.Len() == 0 {
		t.Error("verbose mode produced no log output")
	}
}

// TestSupervisorServeAcceptError verifies that a non-shutdown accept error is
// surfaced as a failure rather than treated as a clean stop.
func TestSupervisorServeAcceptError(t *testing.T) {
	t.Parallel()

	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	s := supervisor{
		sock:     &fakeCloser{},
		errLabel: "accept",
		accept:   func() (net.Conn, bool, error) { return nil, false, errors.New("boom") },
		handler:  func(net.Conn) error { return nil },
	}
	// Context is never cancelled, so the accept error must propagate.
	if err := s.serve(context.Background(), stdio); err == nil {
		t.Fatal("expected accept error to be surfaced as a failure")
	}
}
