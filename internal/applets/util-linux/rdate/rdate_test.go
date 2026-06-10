package rdate

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// serveTime starts a one-shot RFC 868 server returning the given Unix time and
// points the package port at it.
func serveTime(t *testing.T, unixSecs int64) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	orig := port
	port = p
	t.Cleanup(func() { port = orig; _ = ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(unixSecs+epochOffset))
		_, _ = conn.Write(buf[:])
	}()
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestPrintRemoteTime(t *testing.T) {
	const secs = 1700000000
	serveTime(t, secs)
	out, err := run(t, "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Unix(secs, 0).Format("Mon Jan _2 15:04:05 2006") + "\n"
	if out != want {
		t.Errorf("rdate = %q, want %q", out, want)
	}
}

func TestSetClock(t *testing.T) {
	const secs = 1600000000
	serveTime(t, secs)
	var got time.Time
	orig := setTime
	setTime = func(tm time.Time) error { got = tm; return nil }
	defer func() { setTime = orig }()

	if _, err := run(t, "-s", "127.0.0.1"); err != nil {
		t.Fatal(err)
	}
	if got.Unix() != secs {
		t.Errorf("set clock to %d, want %d", got.Unix(), secs)
	}
}

func TestMissingHost(t *testing.T) {
	t.Parallel()
	if _, err := run(t); err == nil {
		t.Errorf("missing host should fail")
	}
}

func TestUnreachable(t *testing.T) {
	// Bind an ephemeral port and immediately close it, so the port is known to
	// be closed rather than assuming a fixed port is free.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close()

	orig := port
	port = p
	defer func() { port = orig }()
	origTO := timeout
	timeout = 500 * time.Millisecond
	defer func() { timeout = origTO }()
	if _, err := run(t, "127.0.0.1"); err == nil {
		t.Errorf("an unreachable host should fail")
	}
}
