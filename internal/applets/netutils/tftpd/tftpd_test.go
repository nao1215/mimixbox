package tftpd

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/netutils/internal/memnet"
	"github.com/nao1215/mimixbox/internal/command"
)

func rrq(name string) []byte {
	p := make([]byte, 2)
	binary.BigEndian.PutUint16(p, opRRQ)
	p = append(p, name...)
	p = append(p, 0)
	p = append(p, "octet"...)
	p = append(p, 0)
	return p
}

func TestParseRequest(t *testing.T) {
	t.Parallel()
	op, name, mode, err := ParseRequest(rrq("file.txt"))
	if err != nil {
		t.Fatalf("ParseRequest error: %v", err)
	}
	if op != opRRQ || name != "file.txt" || mode != "octet" {
		t.Errorf("got op=%d name=%q mode=%q", op, name, mode)
	}
}

func TestSafeJoin(t *testing.T) {
	t.Parallel()
	root := "/srv/tftp"
	// The leading-slash clean confines traversal inside root rather than
	// erroring: "../etc/passwd" resolves to "/srv/tftp/etc/passwd".
	got, err := safeJoin(root, "../etc/passwd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(root, "etc/passwd") {
		t.Errorf("safeJoin traversal = %q, want it confined under root", got)
	}
	got, err = safeJoin(root, "sub/file")
	if err != nil || got != filepath.Join(root, "sub/file") {
		t.Errorf("safeJoin = %q, %v", got, err)
	}
}

// serveOnPipe starts Serve on a server endpoint of an in-memory packet pipe and
// returns the client endpoint plus a stop function that cancels Serve and waits
// for it to return. No real UDP socket is used.
func serveOnPipe(t *testing.T, root string) (*memnet.PacketPipe, func()) {
	t.Helper()
	server, client := memnet.NewPacketPipe()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = Serve(ctx, server, root)
	}()
	return client, func() {
		cancel()
		wg.Wait()
	}
}

func TestServeReadFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	content := bytes.Repeat([]byte("ab"), 400) // 800 bytes -> two blocks
	if err := os.WriteFile(filepath.Join(root, "data.bin"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	client, stop := serveOnPipe(t, root)
	defer stop()

	if _, err := client.WriteTo(rrq("data.bin"), client.PeerAddr()); err != nil {
		t.Fatalf("write: %v", err)
	}

	var got []byte
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	for {
		n, _, err := client.ReadFrom(buf)
		if err != nil {
			break
		}
		if binary.BigEndian.Uint16(buf[0:2]) != opDATA {
			t.Fatalf("expected DATA, got opcode %d", binary.BigEndian.Uint16(buf[0:2]))
		}
		payload := buf[4:n]
		got = append(got, payload...)
		if len(payload) < blockSize {
			break
		}
	}
	if !bytes.Equal(got, content) {
		t.Errorf("transferred %d bytes, want %d", len(got), len(content))
	}
}

func TestServeMissingFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	client, stop := serveOnPipe(t, root)
	defer stop()

	if _, err := client.WriteTo(rrq("nope.txt"), client.PeerAddr()); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 64)
	n, _, err := client.ReadFrom(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if binary.BigEndian.Uint16(buf[0:2]) != opERROR {
		t.Errorf("expected ERROR packet, got opcode %d", binary.BigEndian.Uint16(buf[0:2]))
	}
	_ = n
}

func TestServeWriteRefused(t *testing.T) {
	t.Parallel()
	client, stop := serveOnPipe(t, t.TempDir())
	defer stop()

	// Craft a WRQ; tftpd is read-only and must refuse it with an ERROR.
	wrq := make([]byte, 2)
	binary.BigEndian.PutUint16(wrq, opWRQ)
	wrq = append(wrq, "f.txt"...)
	wrq = append(wrq, 0)
	wrq = append(wrq, "octet"...)
	wrq = append(wrq, 0)
	if _, err := client.WriteTo(wrq, client.PeerAddr()); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 64)
	if _, _, err := client.ReadFrom(buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if binary.BigEndian.Uint16(buf[0:2]) != opERROR || binary.BigEndian.Uint16(buf[2:4]) != 2 {
		t.Errorf("expected ERROR code 2 for WRQ, got op=%d code=%d",
			binary.BigEndian.Uint16(buf[0:2]), binary.BigEndian.Uint16(buf[2:4]))
	}
}

func TestRunValidation(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"somedir"}); err == nil {
		t.Error("expected error without -f")
	}
	if err := New().Run(context.Background(), stdio, []string{"-f"}); err == nil {
		t.Error("expected error without root dir")
	}
}
