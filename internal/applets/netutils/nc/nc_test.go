package nc

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestConnectSendsAndReceives(t *testing.T) {
	// Inject an in-memory dialer: nc gets the client side of a pipe and the
	// server side runs an echo fixture. No loopback socket is used.
	server, client := net.Pipe()
	orig := dial
	dial = func(string, string) (net.Conn, error) { return client, nil }
	t.Cleanup(func() { dial = orig })

	done := make(chan string, 1)
	go func() {
		defer func() { _ = server.Close() }()
		buf := make([]byte, 64)
		n, _ := server.Read(buf)
		_, _ = server.Write([]byte("pong"))
		done <- string(buf[:n])
	}()

	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader("ping"), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"127.0.0.1", "9999"}); err != nil {
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
	// Inject a dialer that fails, mimicking a refused connection without a socket.
	orig := dial
	dial = func(string, string) (net.Conn, error) { return nil, net.UnknownNetworkError("refused") }
	t.Cleanup(func() { dial = orig })

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
	// Inject an in-memory packet conn so serveUDP receives a datagram without a
	// real UDP socket.
	server, client := memnet.NewPacketPipe()
	orig := listenPacket
	listenPacket = func(string, string) (net.PacketConn, error) { return server, nil }
	t.Cleanup(func() { listenPacket = orig })

	out := &bytes.Buffer{}
	errCh := make(chan error, 1)
	go func() {
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		errCh <- New().Run(context.Background(), io, []string{"-u", "-l", "-p", "9000"})
	}()

	// Give the listener a moment to bind to the injected conn, then send.
	time.Sleep(50 * time.Millisecond)
	if _, err := client.WriteTo([]byte("hello-udp"), client.PeerAddr()); err != nil {
		t.Fatalf("write: %v", err)
	}

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
	// Inject an in-memory listener so serve accepts one connection without a
	// real TCP socket.
	ln := memnet.NewPipeListener()
	orig := listen
	listen = func(string, string) (net.Listener, error) { return ln, nil }
	t.Cleanup(func() { listen = orig })

	out := &bytes.Buffer{}
	errCh := make(chan error, 1)
	go func() {
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		errCh <- New().Run(context.Background(), io, []string{"-l", "-p", "9001"})
	}()

	conn, err := ln.Dial()
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, _ = conn.Write([]byte("from-client"))
	_ = conn.(memnet.HalfCloseConn).CloseWrite()

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
