package nc

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestConnectSendsAndReceives(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP/UDP listen unavailable: %v", err)
	}
	defer func() { _ = ln.Close() }()

	done := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- ""
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		_, _ = conn.Write([]byte("pong"))
		done <- string(buf[:n])
	}()

	_, port, _ := net.SplitHostPort(ln.Addr().String())
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader("ping"), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"127.0.0.1", port}); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	select {
	case got := <-done:
		if got != "ping" {
			t.Errorf("server received %q, want ping", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not receive data")
	}
	if out.String() != "pong" {
		t.Errorf("client output = %q, want pong", out.String())
	}
}

func TestListenAddr(t *testing.T) {
	t.Parallel()
	if got, err := listenAddr("8080", nil); err != nil || got != ":8080" {
		t.Errorf("listenAddr(8080) = %q, %v", got, err)
	}
	if got, err := listenAddr("", []string{"9000"}); err != nil || got != ":9000" {
		t.Errorf("listenAddr operand = %q, %v", got, err)
	}
	if got, err := listenAddr("", []string{"127.0.0.1", "9000"}); err != nil || got != "127.0.0.1:9000" {
		t.Errorf("listenAddr host+port = %q, %v", got, err)
	}
	if _, err := listenAddr("", nil); err == nil {
		t.Error("expected error with no port")
	}
}

func TestDialAddr(t *testing.T) {
	t.Parallel()
	if h, p, err := dialAddr("", []string{"example.com", "80"}); err != nil || h != "example.com" || p != "80" {
		t.Errorf("dialAddr = %q %q %v", h, p, err)
	}
	if h, p, err := dialAddr("443", []string{"host"}); err != nil || h != "host" || p != "443" {
		t.Errorf("dialAddr with -p = %q %q %v", h, p, err)
	}
	if _, _, err := dialAddr("", []string{"only-host"}); err == nil {
		t.Error("expected error with host but no port")
	}
	if _, _, err := dialAddr("", nil); err == nil {
		t.Error("expected error with no operands")
	}
}

func TestConnectRefused(t *testing.T) {
	t.Parallel()
	// Port 1 on loopback is virtually always closed.
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"127.0.0.1", "1"}); err == nil {
		t.Fatal("expected a connection error")
	}
}

func TestMissingOperands(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Fatal("expected error with no operands")
	}
}

func TestUDPRoundTrip(t *testing.T) {
	t.Parallel()
	// Listen with the applet, send a datagram with a plain UDP socket.
	out := &bytes.Buffer{}
	errCh := make(chan error, 1)
	ready := make(chan string, 1)

	// Bind a UDP socket first to learn a free port, then hand it to the applet
	// path via address reuse is awkward; instead listen directly here and feed
	// serveUDP through Run using a fixed ephemeral port.
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP/UDP listen unavailable: %v", err)
	}
	addr := pc.LocalAddr().String()
	_ = pc.Close()

	go func() {
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		_, port, _ := net.SplitHostPort(addr)
		ready <- port
		errCh <- New().Run(context.Background(), io, []string{"-u", "-l", "-p", port})
	}()

	port := <-ready
	// Give the listener a moment to bind.
	time.Sleep(200 * time.Millisecond)
	conn, err := net.Dial("udp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = conn.Write([]byte("hello-udp"))
	_ = conn.Close()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("UDP listener did not return")
	}
	if out.String() != "hello-udp" {
		t.Errorf("received %q, want hello-udp", out.String())
	}
}

func TestListenTCP(t *testing.T) {
	t.Parallel()
	// Find a free TCP port, release it, then have the applet listen on it.
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP/UDP listen unavailable: %v", err)
	}
	_, port, _ := net.SplitHostPort(probe.Addr().String())
	_ = probe.Close()

	out := &bytes.Buffer{}
	errCh := make(chan error, 1)
	go func() {
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		errCh <- New().Run(context.Background(), io, []string{"-l", "-p", port})
	}()

	time.Sleep(200 * time.Millisecond)
	conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = conn.Write([]byte("from-client"))
	if tc, ok := conn.(*net.TCPConn); ok {
		_ = tc.CloseWrite()
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not return")
	}
	_ = conn.Close()
	if out.String() != "from-client" {
		t.Errorf("server received %q, want from-client", out.String())
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "nc" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
