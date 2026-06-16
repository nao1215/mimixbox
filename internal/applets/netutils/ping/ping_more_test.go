package ping

import (
	"encoding/binary"
	"testing"
)

// TestParseReplyWithOptions checks that parseReply honors the IHL field when
// the IPv4 header carries options (ihl > 5), locating the ICMP message at the
// correct offset.
func TestParseReplyWithOptions(t *testing.T) {
	t.Parallel()
	// ihl = 6 (24-byte header) => version/ihl byte 0x46.
	const ihl = 6
	packet := make([]byte, ihl*4+8)
	packet[0] = 0x40 | ihl
	icmp := packet[ihl*4:]
	icmp[0] = icmpEchoReply
	binary.BigEndian.PutUint16(icmp[4:], 0x0102)
	binary.BigEndian.PutUint16(icmp[6:], 99)

	id, seq, ok := parseReply(packet)
	if !ok || id != 0x0102 || seq != 99 {
		t.Errorf("parseReply = %#x, %d, %v; want 0x0102, 99, true", id, seq, ok)
	}
}

// TestParseReplyTruncatedAfterIHL covers the second length guard: the packet is
// at least 20 bytes but shorter than ihl*4+8.
func TestParseReplyTruncatedAfterIHL(t *testing.T) {
	t.Parallel()
	// ihl = 15 (max) requires 60+8 bytes; supply far fewer.
	packet := make([]byte, 24)
	packet[0] = 0x4f // ihl = 15
	if _, _, ok := parseReply(packet); ok {
		t.Error("packet too short for its IHL should not parse")
	}
}

// TestChecksumOddLength exercises the trailing-byte branch of checksum (odd
// length) and verifies the one's-complement result.
func TestChecksumOddLength(t *testing.T) {
	t.Parallel()
	// Single byte 0x01 -> sum 0x0100 -> ^0x0100 = 0xfeff.
	if got := checksum([]byte{0x01}); got != 0xfeff {
		t.Errorf("checksum(odd) = %#x, want 0xfeff", got)
	}
}

// TestChecksumCarryFold exercises the high-bit carry fold loop with values that
// overflow 16 bits when summed.
func TestChecksumCarryFold(t *testing.T) {
	t.Parallel()
	// 0xffff + 0xffff = 0x1fffe; folded: 0xfffe + 1 = 0xffff; ^0xffff = 0.
	if got := checksum([]byte{0xff, 0xff, 0xff, 0xff}); got != 0 {
		t.Errorf("checksum(carry) = %#x, want 0", got)
	}
}

// TestEchoRequestSeqWraps confirms a large sequence number is truncated to 16
// bits, matching the on-wire field width.
func TestEchoRequestSeqWraps(t *testing.T) {
	t.Parallel()
	pkt := echoRequest(0x10000, 0x10001)
	if id := binary.BigEndian.Uint16(pkt[4:]); id != 0 {
		t.Errorf("id field = %#x, want 0 (0x10000 truncated)", id)
	}
	if seq := binary.BigEndian.Uint16(pkt[6:]); seq != 1 {
		t.Errorf("seq field = %d, want 1 (0x10001 truncated)", seq)
	}
	// The embedded checksum must still verify to zero.
	if got := checksum(pkt); got != 0 {
		t.Errorf("verification checksum = %#x, want 0", got)
	}
}
