package inetd

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()
	cfg := `# sample
7000 stream tcp nowait root /bin/cat cat -u
9000 dgram udp wait nobody /bin/echo echo hi
`
	svcs, err := ParseConfig(strings.NewReader(cfg))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}
	if len(svcs) != 2 {
		t.Fatalf("got %d services, want 2", len(svcs))
	}
	if svcs[0].Port != 7000 || svcs[0].Protocol != "tcp" || svcs[0].Wait {
		t.Errorf("svc0 = %+v", svcs[0])
	}
	if svcs[0].Program != "/bin/cat" || len(svcs[0].Args) != 2 {
		t.Errorf("svc0 program/args = %q %v", svcs[0].Program, svcs[0].Args)
	}
	if !svcs[1].Wait || svcs[1].Protocol != "udp" {
		t.Errorf("svc1 = %+v", svcs[1])
	}
}

func TestParseConfigErrors(t *testing.T) {
	t.Parallel()
	bad := []string{
		"7000 stream tcp nowait root", // too few fields
		"abc stream tcp nowait root /bin/cat",
		"7000 weird tcp nowait root /bin/cat",
		"7000 stream sctp nowait root /bin/cat",
		"7000 stream tcp maybe root /bin/cat",
	}
	for _, b := range bad {
		if _, err := ParseConfig(strings.NewReader(b)); err == nil {
			t.Errorf("expected error for %q", b)
		}
	}
}

func TestAcceptLoopWithInjectedRunner(t *testing.T) {
	t.Parallel()
	// Drive the accept loop over an in-memory listener: no real socket and no
	// forked process. Serve itself binds real sockets, so the dispatch behavior
	// is exercised here through its acceptLoop helper instead.
	ln := memnet.NewPipeListener()
	svc := Service{Port: 7000, Socket: "stream", Protocol: "tcp", Wait: false, User: "root", Program: "echo"}

	// Injected runner: an upper-casing echo, no real process.
	runner := func(_ context.Context, _ Service, conn net.Conn) error {
		b, _ := io.ReadAll(conn)
		_, _ = conn.Write(bytes.ToUpper(b))
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// acceptLoop relies on its caller (Serve) to close the listener on
	// cancellation; mirror that so the blocked Accept unblocks on shutdown.
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = acceptLoop(ctx, ln, svc, runner)
	}()

	conn, err := ln.Dial()
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, _ = conn.Write([]byte("hi"))
	_ = conn.(memnet.HalfCloseConn).CloseWrite()
	got, _ := io.ReadAll(conn)
	_ = conn.Close()
	if string(got) != "HI" {
		t.Errorf("got %q, want HI", string(got))
	}

	cancel()
	wg.Wait()
}

func TestRunRequiresForegroundAndConfig(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"some.conf"}); err == nil {
		t.Fatal("expected error without -f")
	}
	if err := New().Run(context.Background(), stdio, []string{"-f"}); err == nil {
		t.Fatal("expected error without config path")
	}
}
