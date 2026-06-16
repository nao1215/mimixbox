package ping

import (
	"bytes"
	"context"
	"encoding/binary"
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

func TestEchoRequestChecksumIsValid(t *testing.T) {
	t.Parallel()
	pkt := echoRequest(0x1234, 7)
	if pkt[0] != icmpEchoRequest {
		t.Errorf("type = %d, want %d", pkt[0], icmpEchoRequest)
	}
	// Recomputing the checksum over a packet that already carries it yields 0.
	if got := checksum(pkt); got != 0 {
		t.Errorf("verification checksum = %#x, want 0", got)
	}
	if id := binary.BigEndian.Uint16(pkt[4:]); id != 0x1234 {
		t.Errorf("id = %#x, want 0x1234", id)
	}
	if seq := binary.BigEndian.Uint16(pkt[6:]); seq != 7 {
		t.Errorf("seq = %d, want 7", seq)
	}
}

func TestParseReply(t *testing.T) {
	t.Parallel()
	// Craft an IPv4 header (ihl=5 => 20 bytes) followed by an ICMP echo reply.
	packet := make([]byte, 20+8)
	packet[0] = 0x45 // version 4, ihl 5
	icmp := packet[20:]
	icmp[0] = icmpEchoReply
	binary.BigEndian.PutUint16(icmp[4:], 0xabcd)
	binary.BigEndian.PutUint16(icmp[6:], 42)

	id, seq, ok := parseReply(packet)
	if !ok || id != 0xabcd || seq != 42 {
		t.Errorf("parseReply = %#x, %d, %v", id, seq, ok)
	}
}

func TestParseReplyRejectsNonReply(t *testing.T) {
	t.Parallel()
	packet := make([]byte, 28)
	packet[0] = 0x45
	packet[20] = icmpEchoRequest // not a reply
	if _, _, ok := parseReply(packet); ok {
		t.Error("echo request should not parse as a reply")
	}
	if _, _, ok := parseReply([]byte{0x45}); ok {
		t.Error("short packet should not parse")
	}
}

func TestChecksumKnown(t *testing.T) {
	t.Parallel()
	// One's-complement sum of 0x0000 is 0xffff.
	if got := checksum([]byte{0x00, 0x00}); got != 0xffff {
		t.Errorf("checksum = %#x, want 0xffff", got)
	}
}

func TestMissingHost(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exactly one HOST") {
		t.Errorf("err = %v", err)
	}
}

func TestUnresolvableHost(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "no-such-host.invalid.")
	if err == nil {
		t.Fatal("expected resolve error")
	}
	if !strings.Contains(err.Error(), "cannot resolve") {
		t.Errorf("err = %v", err)
	}
}

func TestPingLoopback(t *testing.T) {
	t.Parallel()
	// Exercises ping(): as non-root the raw socket fails with a privilege
	// error; as root it pings loopback and returns quickly. Either is fine.
	out, _, err := run(t, "-c", "1", "-i", "0", "-W", "1", "127.0.0.1")
	if err != nil {
		if !strings.Contains(err.Error(), "raw ICMP socket") {
			t.Errorf("unexpected error = %v", err)
		}
		return
	}
	// Ran as root: we should have printed the PING banner and statistics.
	if !strings.Contains(out, "PING 127.0.0.1") || !strings.Contains(out, "statistics") {
		t.Errorf("out = %q", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "ping" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
