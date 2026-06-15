package tftp

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestPacketHelpers(t *testing.T) {
	t.Parallel()
	rrq := request(opRRQ, "file.bin")
	if binary.BigEndian.Uint16(rrq[0:2]) != opRRQ {
		t.Error("RRQ opcode wrong")
	}
	if !bytes.Contains(rrq, []byte("octet")) {
		t.Error("RRQ should use octet mode")
	}
	a := ack(5)
	op, block, _, _ := parsePacket(a)
	if op != opACK || block != 5 {
		t.Errorf("ack roundtrip: op=%d block=%d", op, block)
	}
	d := dataPacket(2, []byte("hello"))
	op, block, payload, _ := parsePacket(d)
	if op != opDATA || block != 2 || string(payload) != "hello" {
		t.Errorf("data roundtrip: op=%d block=%d payload=%q", op, block, payload)
	}
}

// startServer runs a tiny single-transfer TFTP server on loopback. mode "get"
// serves content for an RRQ; mode "put" stores into *stored for a WRQ. The
// returned channel is closed once the server goroutine has finished (and, for
// "put", after *stored has been assigned), giving callers a happens-before edge
// so they can read *stored without racing the goroutine.
func startServer(t *testing.T, mode string, content []byte, stored *[]byte) (string, <-chan struct{}) {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback UDP unavailable: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4+blockSize)
		n, client, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		op := int(binary.BigEndian.Uint16(buf[0:2]))
		switch {
		case mode == "get" && op == opRRQ:
			serveGet(pc, client, content)
		case mode == "put" && op == opWRQ:
			servePut(t, pc, client, stored)
		}
		_ = n
	}()
	return pc.LocalAddr().String(), done
}

func serveGet(pc net.PacketConn, client net.Addr, content []byte) {
	block := uint16(1)
	for off := 0; ; off += blockSize {
		end := off + blockSize
		if end > len(content) {
			end = len(content)
		}
		chunk := content[off:end]
		_, _ = pc.WriteTo(dataPacket(block, chunk), client)
		// Wait for the ACK.
		ackBuf := make([]byte, 4)
		_, _, _ = pc.ReadFrom(ackBuf)
		block++
		if len(chunk) < blockSize {
			return
		}
	}
}

func servePut(t *testing.T, pc net.PacketConn, client net.Addr, stored *[]byte) {
	t.Helper()
	// ACK 0 to start.
	_, _ = pc.WriteTo(ack(0), client)
	var got bytes.Buffer
	for {
		buf := make([]byte, 4+blockSize)
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			return
		}
		op, block, payload, _ := parsePacket(buf[:n])
		if op != opDATA {
			return
		}
		got.Write(payload)
		_, _ = pc.WriteTo(ack(block), client)
		if len(payload) < blockSize {
			*stored = got.Bytes()
			return
		}
	}
}

func TestGet(t *testing.T) {
	content := bytes.Repeat([]byte("A"), 700) // spans two blocks
	addr, _ := startServer(t, "get", content, nil)
	host, port, _ := net.SplitHostPort(addr)

	var written []byte
	orig := writeLocal
	writeLocal = func(_ string, data []byte, _ os.FileMode) error { written = data; return nil }
	t.Cleanup(func() { writeLocal = orig })

	out, _, err := run(t, "-g", "-l", "out.bin", "-r", "file.bin", host, port)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !bytes.Equal(written, content) {
		t.Errorf("downloaded %d bytes, want %d", len(written), len(content))
	}
	if !strings.Contains(out, "Received 700 bytes") {
		t.Errorf("out = %q", out)
	}
}

func TestPut(t *testing.T) {
	content := bytes.Repeat([]byte("B"), 512) // exactly one full block then empty
	var stored []byte
	addr, done := startServer(t, "put", nil, &stored)
	host, port, _ := net.SplitHostPort(addr)

	origR := readLocal
	readLocal = func(string) ([]byte, error) { return content, nil }
	t.Cleanup(func() { readLocal = origR })

	out, _, err := run(t, "-p", "-l", "in.bin", "-r", "file.bin", host, port)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	<-done // wait for the server goroutine to finish writing *stored
	if !bytes.Equal(stored, content) {
		t.Errorf("server stored %d bytes, want %d", len(stored), len(content))
	}
	if !strings.Contains(out, "Sent 512 bytes") {
		t.Errorf("out = %q", out)
	}
}

func TestArgValidation(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, "-l", "x", "-r", "y", "127.0.0.1"); err == nil {
		t.Error("expected error without -g or -p")
	}
	if _, _, err := run(t, "-g", "-p", "-l", "x", "-r", "y", "127.0.0.1"); err == nil {
		t.Error("expected error with both -g and -p")
	}
	if _, _, err := run(t, "-g", "-r", "y", "127.0.0.1"); err == nil {
		t.Error("expected error without -l")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "tftp" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
