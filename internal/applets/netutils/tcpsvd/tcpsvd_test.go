package tcpsvd

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestServeTCPDispatchesConnections(t *testing.T) {
	t.Parallel()
	// In-memory listener: the TCP accept loop runs without a real socket.
	ln := memnet.NewPipeListener()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	// Echo handler: read input and write it back.
	handler := func(conn net.Conn) error {
		b, _ := io.ReadAll(conn)
		_, _ = conn.Write(b)
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = ServeTCP(ctx, ln, stdio, true, handler)
	}()

	conn, err := ln.Dial()
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, _ = conn.Write([]byte("hello"))
	_ = conn.(memnet.HalfCloseConn).CloseWrite()
	got, _ := io.ReadAll(conn)
	_ = conn.Close()
	if string(got) != "hello" {
		t.Errorf("echo = %q, want hello", string(got))
	}

	cancel()
	wg.Wait()
}

func TestServeUDPDispatchesDatagrams(t *testing.T) {
	t.Parallel()
	// In-memory packet pipe: the UDP receive loop runs without a real socket.
	server, client := memnet.NewPacketPipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	handler := func(conn net.Conn) error {
		b, _ := io.ReadAll(conn)
		_, _ = conn.Write(bytes.ToUpper(b))
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = ServeUDP(ctx, server, stdio, false, handler)
	}()

	if _, err := client.WriteTo([]byte("ping"), client.PeerAddr()); err != nil {
		t.Fatalf("write udp: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 64)
	n, _, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	if string(buf[:n]) != "PING" {
		t.Errorf("reply = %q, want PING", string(buf[:n]))
	}

	cancel()
	wg.Wait()
}

func TestRunRejectsMissingArgs(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewTcpsvd().Run(context.Background(), stdio, []string{"127.0.0.1"}); err == nil {
		t.Fatal("expected error for missing PORT/PROG")
	}
	if err := NewUdpsvd().Run(context.Background(), stdio, []string{"127.0.0.1", "notaport", "cat"}); err == nil {
		t.Fatal("expected error for invalid port")
	}
}
