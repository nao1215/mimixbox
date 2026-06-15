package ntpd

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestNTPTimestampRoundTrip(t *testing.T) {
	t.Parallel()
	in := time.Unix(1700000000, 500000000).UTC()
	out := FromNTP(ToNTP(in))
	if d := out.Sub(in); d < -time.Millisecond || d > time.Millisecond {
		t.Errorf("round trip drift = %v", d)
	}
}

func TestPacketRoundTrip(t *testing.T) {
	t.Parallel()
	p := &Packet{LeapVersionMode: 0x23, Stratum: 2, TransmitTime: 0x1122334455667788}
	got, err := Unmarshal(p.Marshal())
	if err != nil {
		t.Fatal(err)
	}
	if got.TransmitTime != p.TransmitTime || got.LeapVersionMode != p.LeapVersionMode {
		t.Errorf("round trip lost data: %+v", got)
	}
}

// fixedServer answers with a server reply offset by a known amount.
type fixedServer struct {
	serverTime time.Time
}

func (f *fixedServer) Query(request []byte) ([]byte, error) {
	req, err := Unmarshal(request)
	if err != nil {
		return nil, err
	}
	reply := ServerReply(req, f.serverTime, f.serverTime, 2)
	return reply.Marshal(), nil
}

func TestOffsetCalculation(t *testing.T) {
	t.Parallel()
	now := time.Unix(1000, 0).UTC()
	// Server clock is 5 seconds ahead of the client.
	server := &fixedServer{serverTime: now.Add(5 * time.Second)}
	offset, reply, err := Offset(func() time.Time { return now }, server)
	if err != nil {
		t.Fatalf("Offset error: %v", err)
	}
	if reply.Stratum != 2 {
		t.Errorf("stratum = %d, want 2", reply.Stratum)
	}
	// Offset should be close to +5s (allow slack for the local t4 read).
	if offset < 4*time.Second || offset > 6*time.Second {
		t.Errorf("offset = %v, want ~5s", offset)
	}
}

func TestRunCapabilityGated(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), stdio, []string{"-q", "-p", "time.example.com"})
	if err == nil || !strings.Contains(err.Error(), "capability-gated") {
		t.Errorf("expected documented capability error, got %v", err)
	}
	if err := New().Run(context.Background(), stdio, []string{"-q"}); err == nil {
		t.Error("expected error when server is missing")
	}
}
