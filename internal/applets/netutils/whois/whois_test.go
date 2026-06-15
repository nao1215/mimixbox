package whois

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, fn func(server, object string) (string, error)) {
	t.Helper()
	orig := query
	query = fn
	t.Cleanup(func() { query = orig })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestQueryDefaultServer(t *testing.T) {
	var gotServer, gotObject string
	stub(t, func(server, object string) (string, error) {
		gotServer, gotObject = server, object
		return "Domain Name: EXAMPLE.TEST\n", nil
	})
	out, _, err := run(t, "example.test")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if gotServer != defaultServer {
		t.Errorf("server = %q, want %q", gotServer, defaultServer)
	}
	if gotObject != "example.test" {
		t.Errorf("object = %q", gotObject)
	}
	if !strings.Contains(out, "Domain Name: EXAMPLE.TEST") {
		t.Errorf("response not printed: %s", out)
	}
}

func TestQueryCustomServer(t *testing.T) {
	var gotServer string
	stub(t, func(server, _ string) (string, error) {
		gotServer = server
		return "ok\n", nil
	})
	if _, _, err := run(t, "-h", "whois.example.test", "example.test"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if gotServer != "whois.example.test" {
		t.Errorf("server = %q, want whois.example.test", gotServer)
	}
}

func TestQueryFailure(t *testing.T) {
	stub(t, func(string, string) (string, error) { return "", errors.New("connection refused") })
	if _, _, err := run(t, "example.test"); err == nil {
		t.Error("expected error on query failure")
	}
}

func TestBadArgs(t *testing.T) {
	stub(t, func(string, string) (string, error) { return "", nil })
	if _, _, err := run(t); err == nil {
		t.Error("expected error with no operand")
	}
	if _, _, err := run(t, "a", "b"); err == nil {
		t.Error("expected error with two operands")
	}
}

// TestTCPQueryAgainstLocalServer exercises the real protocol path against a
// loopback stub so the wire format is covered without the public internet.
func TestTCPQueryAgainstLocalServer(t *testing.T) {
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
		buf := make([]byte, 128)
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte("Domain Name: EXAMPLE.TEST\nRegistrar: Test\n"))
	}()

	resp, err := tcpQuery(ln.Addr().String(), "example.test")
	if err != nil {
		t.Fatalf("tcpQuery error = %v", err)
	}
	if !strings.Contains(resp, "EXAMPLE.TEST") {
		t.Errorf("response = %q", resp)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "whois" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
