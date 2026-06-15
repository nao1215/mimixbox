package telnet

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestHostPort(t *testing.T) {
	t.Parallel()
	h, p, err := hostPort([]string{"example.test"})
	if err != nil || h != "example.test" || p != "23" {
		t.Errorf("default port: %q %q %v", h, p, err)
	}
	h, p, err = hostPort([]string{"example.test", "25"})
	if err != nil || p != "25" {
		t.Errorf("explicit port: %q %q %v", h, p, err)
	}
	if _, _, err := hostPort(nil); err == nil {
		t.Error("expected error with no operands")
	}
}

func TestSessionAgainstLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP listen unavailable: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		_, _ = conn.Write([]byte("echo:" + string(buf[:n])))
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	out, _, err := run(t, "hello\n", host, port)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "echo:hello") {
		t.Errorf("out = %q", out)
	}
}

func TestConnectFailure(t *testing.T) {
	orig := dial
	dial = func(string) (net.Conn, error) { return nil, net.UnknownNetworkError("nope") }
	t.Cleanup(func() { dial = orig })
	if _, _, err := run(t, "", "127.0.0.1", "1"); err == nil {
		t.Error("expected connection error")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "telnet" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
