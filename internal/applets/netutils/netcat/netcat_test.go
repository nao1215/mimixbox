package netcat

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}
	err := New().Run(context.Background(), stdio, args)
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
//
// INTEGRATION SUBSET: this is the one netutils test that still requires a real
// loopback socket. netcat is a thin alias that forwards to the nc applet, whose
// network seams (dial/listen/listenPacket) are unexported, so they cannot be
// injected from this package. The byte-shuttling logic itself is covered fully
// in-memory by the nc package's own tests (see internal/.../nc/nc_test.go);
// here we only assert that netcat delegates to that path end to end, which is
// why a genuine socket is used. It skips when loopback is unavailable.
func TestDelegatesToNc(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("integration subset: loopback TCP listen unavailable: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		// A deadline keeps the drain from blocking forever if the client never
		// closes its side.
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		_, _ = conn.Write([]byte("pong"))
		// Drain whatever the client sends before closing. Closing with unread
		// bytes still buffered makes the kernel send an RST, which the client
		// surfaces as "connection reset by peer" and made this test flaky.
		_, _ = io.Copy(io.Discard, conn)
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

// TestHelpSections asserts `netcat --help` renders netcat-named structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("--help err = %v", err)
	}
	for _, want := range []string{"Usage: netcat", "Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help missing %q: %q", want, out.String())
		}
	}
}
