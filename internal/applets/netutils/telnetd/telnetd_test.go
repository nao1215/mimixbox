package telnetd

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func TestStripIAC(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"plain", []byte("hello"), []byte("hello")},
		{"do-option", []byte{'h', iac, 253, 24, 'i'}, []byte("hi")},
		{"escaped-ff", []byte{'a', iac, iac, 'b'}, []byte{'a', 0xff, 'b'}},
		{"subneg", []byte{'x', iac, sb, 24, 1, iac, se, 'z'}, []byte("xz")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripIAC(tt.in); !bytes.Equal(got, tt.want) {
				t.Errorf("StripIAC(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestReaderAcrossBoundaries(t *testing.T) {
	t.Parallel()
	// Feed the IAC sequence split across two reads.
	src := &chunkReader{chunks: [][]byte{
		{'a', iac},
		{253, 24, 'b'},
	}}
	r := NewReader(src)
	got, _ := io.ReadAll(r)
	if string(got) != "ab" {
		t.Errorf("Reader output = %q, want ab", string(got))
	}
}

// chunkReader returns its chunks one Read at a time, then io.EOF.
type chunkReader struct {
	chunks [][]byte
	i      int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if c.i >= len(c.chunks) {
		return 0, io.EOF
	}
	n := copy(p, c.chunks[c.i])
	c.i++
	return n, nil
}

func TestServeRunsHandler(t *testing.T) {
	t.Parallel()
	// In-memory listener: the accept loop runs without a real socket. The
	// connections support CloseWrite, so the handler's ReadAll sees EOF and
	// can still write its reply on the other half.
	ln := memnet.NewPipeListener()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Session handler reads IAC-filtered input and echoes it uppercased.
	handler := func(conn net.Conn) error {
		b, _ := io.ReadAll(NewReader(conn))
		_, _ = conn.Write(bytes.ToUpper(b))
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = Serve(ctx, ln, handler)
	}()

	conn, err := ln.Dial()
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	_, _ = conn.Write([]byte{'h', iac, 253, 24, 'i'})
	_ = conn.(memnet.HalfCloseConn).CloseWrite()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	got, _ := io.ReadAll(conn)
	_ = conn.Close()
	if string(got) != "HI" {
		t.Errorf("session echo = %q, want HI", string(got))
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
