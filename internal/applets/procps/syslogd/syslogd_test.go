package syslogd

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// waitFor polls cond until it is true or the deadline passes.
func waitFor(cond func() bool) bool {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestReceivesAndLogs(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "log.sock")
	logfile := filepath.Join(dir, "messages")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
		done <- New().Run(ctx, io, []string{"-l", sock, "-O", logfile})
	}()

	if !waitFor(func() bool { _, err := os.Stat(sock); return err == nil }) {
		cancel()
		t.Fatal("syslogd never created its socket")
	}

	conn, err := net.Dial("unixgram", sock)
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	if _, err := conn.Write([]byte("<13>hello world")); err != nil {
		cancel()
		t.Fatalf("write: %v", err)
	}
	_ = conn.Close()

	logged := waitFor(func() bool {
		data, _ := os.ReadFile(logfile)
		return strings.Contains(string(data), "hello world")
	})

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("syslogd did not stop after cancellation")
	}

	if !logged {
		data, _ := os.ReadFile(logfile)
		t.Fatalf("message not logged; file = %q", data)
	}
	// The priority prefix must be stripped.
	data, _ := os.ReadFile(logfile)
	if strings.Contains(string(data), "<13>") {
		t.Errorf("priority prefix not stripped: %q", data)
	}
}

func TestStripPriority(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"<13>message":    "message",
		"no prefix":      "no prefix",
		"<7>trailing\n":  "trailing",
		"<not a number>": "<not a number>",
	}
	for in, want := range cases {
		if got := stripPriority(in); got != want {
			t.Errorf("stripPriority(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBadSocketFails(t *testing.T) {
	t.Parallel()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, []string{"-l", "/no/such/dir/log.sock", "-O", "/tmp/x"})
	if err == nil {
		t.Errorf("an unbindable socket should fail")
	}
}
