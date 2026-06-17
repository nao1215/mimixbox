package fakeidentd

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestReply(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, query, user, want string
	}{
		{"valid", "6113, 12345\r\n", "alice", "6113 , 12345 : USERID : UNIX : alice\r\n"},
		{"malformed", "garbage\r\n", "bob", "garbage : ERROR : INVALID-PORT\r\n"},
		{"badport", "0,99999\r\n", "bob", "0,99999 : ERROR : INVALID-PORT\r\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reply(tt.query, tt.user); got != tt.want {
				t.Errorf("Reply(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestServeAnswersQuery(t *testing.T) {
	t.Parallel()
	// In-memory listener: Serve's accept loop runs without a real socket.
	ln := memnet.NewPipeListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = Serve(ctx, ln, "carol")
	}()

	conn, err := ln.Dial()
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	_, _ = conn.Write([]byte("113, 5000\r\n"))
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	resp, _ := bufio.NewReader(conn).ReadString('\n')
	if !strings.Contains(resp, "USERID : UNIX : carol") {
		t.Errorf("response = %q", resp)
	}

	cancel()
	wg.Wait()
}

func TestRunRequiresForeground(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"-b", "127.0.0.1:0"}); err == nil {
		t.Fatal("expected error without -f")
	}
}
