package ftpd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestSplitCommand(t *testing.T) {
	t.Parallel()
	v, a := SplitCommand("retr file.txt\r\n")
	if v != "RETR" || a != "file.txt" {
		t.Errorf("got %q %q", v, a)
	}
	v, a = SplitCommand("PWD")
	if v != "PWD" || a != "" {
		t.Errorf("got %q %q", v, a)
	}
}

func TestResolvePath(t *testing.T) {
	t.Parallel()
	tests := []struct{ cwd, arg, want string }{
		{"/", "sub", "/sub"},
		{"/a/b", "..", "/a"},
		{"/a", "/x/y", "/x/y"},
		{"/a", "../../../etc", "/etc"}, // cannot escape root
		{"/x", "", "/x"},
	}
	for _, tt := range tests {
		if got := ResolvePath(tt.cwd, tt.arg); got != tt.want {
			t.Errorf("ResolvePath(%q,%q) = %q, want %q", tt.cwd, tt.arg, got, tt.want)
		}
	}
}

func TestFTPRetrieveOverLoopback(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	want := "file contents here\n"
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback listen unavailable: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = Serve(ctx, ln, root)
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	r := bufio.NewReader(conn)

	readReply := func() string {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		line, _ := r.ReadString('\n')
		return line
	}
	send := func(s string) { _, _ = fmt.Fprintf(conn, "%s\r\n", s) }

	_ = readReply() // 220 banner
	send("USER anonymous")
	_ = readReply()
	send("PASS x@x")
	if !strings.HasPrefix(readReply(), "230") {
		t.Fatal("login failed")
	}
	send("PASV")
	pasv := readReply()
	dataAddr := parsePASV(t, pasv)

	send("RETR hello.txt")
	dataConn, err := net.Dial("tcp", dataAddr)
	if err != nil {
		t.Fatalf("data dial: %v", err)
	}
	if !strings.HasPrefix(readReply(), "150") {
		t.Fatal("expected 150 before transfer")
	}
	got, _ := io.ReadAll(dataConn)
	_ = dataConn.Close()
	if string(got) != want {
		t.Errorf("RETR = %q, want %q", string(got), want)
	}
	if !strings.HasPrefix(readReply(), "226") {
		t.Fatal("expected 226 after transfer")
	}

	send("STOR x")
	if !strings.HasPrefix(readReply(), "550") {
		t.Error("STOR should be refused with 550")
	}
	send("QUIT")
	_ = readReply()

	cancel()
	wg.Wait()
}

// parsePASV extracts the host:port from a 227 PASV reply.
func parsePASV(t *testing.T, line string) string {
	t.Helper()
	open := strings.IndexByte(line, '(')
	close := strings.IndexByte(line, ')')
	if open < 0 || close < 0 {
		t.Fatalf("bad PASV reply: %q", line)
	}
	nums := strings.Split(line[open+1:close], ",")
	if len(nums) != 6 {
		t.Fatalf("bad PASV tuple: %q", line)
	}
	p1, _ := strconv.Atoi(nums[4])
	p2, _ := strconv.Atoi(nums[5])
	port := p1<<8 | p2
	return net.JoinHostPort(strings.Join(nums[:4], "."), strconv.Itoa(port))
}

func TestRunRequiresForeground(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"."}); err == nil {
		t.Fatal("expected error without -f")
	}
}
