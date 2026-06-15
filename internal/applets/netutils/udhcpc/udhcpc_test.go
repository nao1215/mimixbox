package udhcpc

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/netutils/udhcpd"
	"github.com/nao1215/mimixbox/internal/command"
)

// serverTransport pairs the client with an in-memory udhcpd server so a full
// DISCOVER/OFFER/REQUEST/ACK exchange can be unit-tested.
type serverTransport struct {
	cfg   *udhcpd.Config
	alloc *udhcpd.Allocator
	reply []byte
}

func (s *serverTransport) Send(packet []byte) error {
	req, err := udhcpd.Unmarshal(packet)
	if err != nil {
		return err
	}
	rep := udhcpd.HandlePacket(req, s.cfg, s.alloc)
	if rep != nil {
		s.reply = rep.Marshal()
	}
	return nil
}

func (s *serverTransport) Recv() ([]byte, error) {
	return s.reply, nil
}

func TestAcquireFullExchange(t *testing.T) {
	t.Parallel()
	cfg := &udhcpd.Config{
		Start:    net.ParseIP("10.1.1.100").To4(),
		End:      net.ParseIP("10.1.1.110").To4(),
		ServerID: net.ParseIP("10.1.1.1").To4(),
		LeaseSec: 1200,
	}
	tr := &serverTransport{cfg: cfg, alloc: udhcpd.NewAllocator(cfg.Start, cfg.End)}
	mac, _ := net.ParseMAC("12:34:56:78:9a:bc")

	lease, err := Acquire(42, mac, tr)
	if err != nil {
		t.Fatalf("Acquire error: %v", err)
	}
	if !lease.IP.Equal(net.ParseIP("10.1.1.100")) {
		t.Errorf("leased IP = %v, want 10.1.1.100", lease.IP)
	}
	if lease.LeaseSec != 1200 {
		t.Errorf("lease time = %d, want 1200", lease.LeaseSec)
	}
	if !lease.ServerID.Equal(net.ParseIP("10.1.1.1")) {
		t.Errorf("server id = %v", lease.ServerID)
	}
}

func TestRunCapabilityGated(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	err := NewUdhcpc().Run(context.Background(), stdio, []string{"-i", "eth0"})
	if err == nil || !strings.Contains(err.Error(), "capability-gated") {
		t.Errorf("expected documented capability error, got %v", err)
	}
	if err := NewUdhcpc().Run(context.Background(), stdio, nil); err == nil {
		t.Error("expected error when interface is missing")
	}
	if err := NewUdhcpc6().Run(context.Background(), stdio, []string{"-i", "eth0"}); err == nil {
		t.Error("udhcpc6 should also be capability-gated")
	}
}
