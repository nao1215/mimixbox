package netcat

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

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "netcat" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestMissingOperands(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, ""); err == nil {
		t.Error("expected error when HOST and PORT are missing")
	}
}

// TestDelegatesToNc connects netcat to a loopback TCP echo and confirms the
// payload is shuttled through the nc backend.
func TestDelegatesToNc(t *testing.T) {
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
		_, _ = conn.Write([]byte("pong"))
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	out, _, err := run(t, "ping", host, port)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "pong") {
		t.Errorf("out = %q, want it to contain pong", out)
	}
}
