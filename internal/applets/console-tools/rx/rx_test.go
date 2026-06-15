package rx

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// makePacket builds one XMODEM SOH packet for the given block number and data
// (padded/truncated to 128 bytes).
func makePacket(num byte, data []byte) []byte {
	buf := make([]byte, blockSize)
	copy(buf, data)
	for i := len(data); i < blockSize; i++ {
		buf[i] = 0x1a // SUB padding
	}
	pkt := []byte{soh, num, 255 - num}
	pkt = append(pkt, buf...)
	pkt = append(pkt, checksum(buf))
	return pkt
}

func TestReceiveSingleBlock(t *testing.T) {
	t.Parallel()
	var link bytes.Buffer
	link.Write(makePacket(1, []byte("hello xmodem")))
	link.WriteByte(eot)

	var control, out bytes.Buffer
	if err := Receive(context.Background(), &link, &control, &out); err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !strings.HasPrefix(out.String(), "hello xmodem") {
		t.Errorf("out = %q", out.String()[:20])
	}
	if len(out.Bytes()) != blockSize {
		t.Errorf("out length = %d, want %d", len(out.Bytes()), blockSize)
	}
	// Control bytes: initial NAK then an ACK per block and EOT.
	c := control.Bytes()
	if len(c) == 0 || c[0] != nak {
		t.Errorf("first control byte = %#x, want NAK", c[0])
	}
}

func TestReceiveTwoBlocks(t *testing.T) {
	t.Parallel()
	var link bytes.Buffer
	link.Write(makePacket(1, []byte("AAA")))
	link.Write(makePacket(2, []byte("BBB")))
	link.WriteByte(eot)

	var control, out bytes.Buffer
	if err := Receive(context.Background(), &link, &control, &out); err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if out.Len() != 2*blockSize {
		t.Errorf("out length = %d", out.Len())
	}
	if out.Bytes()[0] != 'A' || out.Bytes()[blockSize] != 'B' {
		t.Error("block contents in wrong order")
	}
}

func TestReceiveChecksumError(t *testing.T) {
	t.Parallel()
	pkt := makePacket(1, []byte("data"))
	pkt[len(pkt)-1] ^= 0xff // corrupt checksum
	var link bytes.Buffer
	link.Write(pkt)
	var control, out bytes.Buffer
	if err := Receive(context.Background(), &link, &control, &out); err == nil {
		t.Error("expected checksum error")
	}
}

func TestReceiveComplementError(t *testing.T) {
	t.Parallel()
	pkt := makePacket(1, []byte("data"))
	pkt[2] ^= 0xff // corrupt block-number complement
	var link bytes.Buffer
	link.Write(pkt)
	var control, out bytes.Buffer
	if err := Receive(context.Background(), &link, &control, &out); err == nil {
		t.Error("expected complement mismatch error")
	}
}

func TestReceiveCancel(t *testing.T) {
	t.Parallel()
	var link bytes.Buffer
	link.WriteByte(can)
	var control, out bytes.Buffer
	if err := Receive(context.Background(), &link, &control, &out); err == nil {
		t.Error("expected cancel error")
	}
}

func TestRunNoFile(t *testing.T) {
	t.Parallel()
	var out, errBuf bytes.Buffer
	io := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), io, nil); err == nil {
		t.Error("expected error when no file given")
	}
}

func TestRunReceivesToFile(t *testing.T) {
	t.Parallel()
	var link bytes.Buffer
	link.Write(makePacket(1, []byte("payload")))
	link.WriteByte(eot)

	path := filepath.Join(t.TempDir(), "out.bin")
	var control, errBuf bytes.Buffer
	io := command.IO{In: &link, Out: &control, Err: &errBuf}
	if err := New().Run(context.Background(), io, []string{path}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), "payload") {
		t.Errorf("file content = %q", string(got)[:20])
	}
}
