package tcpsvd

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/nao1215/mimixbox/internal/command"
)

// ConnHandler handles one accepted connection (TCP) or datagram exchange (UDP).
type ConnHandler func(conn net.Conn) error

// acceptFunc produces the next connection to service, blocking until one is
// available. The TCP and UDP loops differ only in how a connection is sourced
// (Listener.Accept versus reading a datagram), so the shared supervisor core
// drives this function and is otherwise protocol-agnostic.
//
// ownsConn reports whether the supervisor should close the connection after the
// handler returns: stream connections (TCP) are owned and serviced concurrently
// by the supervisor, while a per-datagram UDP connection manages its own
// lifetime and is serviced inline.
type acceptFunc func() (conn net.Conn, ownsConn bool, err error)

// supervisor holds the protocol-agnostic accept-loop configuration shared by
// the TCP and UDP servers.
type supervisor struct {
	sock     io.Closer                  // socket closed on ctx cancellation to unblock accept
	verbose  bool                       // log each connection to stderr
	logLine  func(conn net.Conn) string // verbose log line for an accepted connection
	errLabel string                     // prefix for the accept/read failure message
	accept   acceptFunc                 // sources the next connection
	handler  ConnHandler                // services each connection
}

// serve runs the shared accept loop until ctx is cancelled: it closes the
// socket on cancellation, repeatedly sources connections via accept, logs each
// (when verbose), and dispatches handler. Owned (stream) connections run
// concurrently and are awaited on shutdown; unowned (datagram) connections run
// inline.
func (s supervisor) serve(ctx context.Context, stdio command.IO) error {
	go func() {
		<-ctx.Done()
		_ = s.sock.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, ownsConn, err := s.accept()
		if err != nil {
			if ctx.Err() != nil {
				wg.Wait()
				return nil
			}
			return command.Failuref("%s: %v", s.errLabel, err)
		}
		if s.verbose {
			_, _ = fmt.Fprintf(stdio.Err, "%s\n", s.logLine(conn))
		}
		if ownsConn {
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer func() { _ = c.Close() }()
				_ = s.handler(c)
			}(conn)
			continue
		}
		_ = s.handler(conn)
	}
}
