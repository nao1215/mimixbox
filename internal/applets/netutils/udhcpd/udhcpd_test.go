package udhcpd

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	mac, _ := net.ParseMAC("00:11:22:33:44:55")
	m := &Message{
		Op:     OpBootRequest,
		XID:    0xdeadbeef,
		CHAddr: mac,
		Options: map[byte][]byte{
			OptMessageType: {Discover},
			OptRequestedIP: net.ParseIP("192.168.0.50").To4(),
		},
	}
	got, err := Unmarshal(m.Marshal())
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if got.XID != 0xdeadbeef || got.Type() != Discover {
		t.Errorf("round trip lost data: %+v", got)
	}
	if !net.IP(got.Options[OptRequestedIP]).Equal(net.ParseIP("192.168.0.50")) {
		t.Errorf("requested IP option lost: %v", got.Options[OptRequestedIP])
	}
}

func TestParseConfig(t *testing.T) {
	t.Parallel()
	cfg, err := ParseConfig(strings.NewReader(`
# pool
interface eth0
start 192.168.0.20
end 192.168.0.30
server_id 192.168.0.1
lease 3600
opt subnet 255.255.255.0
opt router 192.168.0.1
opt dns 8.8.8.8
`))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}
	if cfg.Interface != "eth0" || cfg.LeaseSec != 3600 {
		t.Errorf("cfg = %+v", cfg)
	}
	if !cfg.Start.Equal(net.ParseIP("192.168.0.20")) || !cfg.End.Equal(net.ParseIP("192.168.0.30")) {
		t.Errorf("pool = %v-%v", cfg.Start, cfg.End)
	}
	if !cfg.DNS.Equal(net.ParseIP("8.8.8.8")) {
		t.Errorf("dns = %v", cfg.DNS)
	}
}

func TestParseConfigErrors(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"start notanip\nend 1.2.3.4", "start 1.2.3.4"} {
		if _, err := ParseConfig(strings.NewReader(in)); err == nil {
			t.Errorf("expected error for %q", in)
		}
	}
}

func TestAllocator(t *testing.T) {
	t.Parallel()
	a := NewAllocator(net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2"))
	got1 := a.Allocate()
	got2 := a.Allocate()
	got3 := a.Allocate()
	if !got1.Equal(net.ParseIP("10.0.0.1")) || !got2.Equal(net.ParseIP("10.0.0.2")) {
		t.Errorf("alloc = %v %v", got1, got2)
	}
	if got3 != nil {
		t.Errorf("pool should be exhausted, got %v", got3)
	}
}

// memTransport is an in-memory Transport for unit tests.
type memTransport struct {
	in   [][]byte
	idx  int
	sent [][]byte
}

func (m *memTransport) Recv() ([]byte, net.Addr, error) {
	if m.idx >= len(m.in) {
		return nil, nil, io.EOF
	}
	p := m.in[m.idx]
	m.idx++
	return p, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 68}, nil
}

func (m *memTransport) Send(p []byte, _ net.Addr) error {
	m.sent = append(m.sent, append([]byte{}, p...))
	return nil
}

func TestServeWithInjectedTransport(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Start:    net.ParseIP("192.168.0.10").To4(),
		End:      net.ParseIP("192.168.0.20").To4(),
		ServerID: net.ParseIP("192.168.0.1").To4(),
		LeaseSec: 600,
	}
	mac, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	discover := (&Message{Op: OpBootRequest, XID: 1, CHAddr: mac, Options: map[byte][]byte{OptMessageType: {Discover}}}).Marshal()
	tr := &memTransport{in: [][]byte{discover}}

	if err := Serve(context.Background(), cfg, tr); err != nil {
		t.Fatalf("Serve error: %v", err)
	}
	if len(tr.sent) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(tr.sent))
	}
	reply, err := Unmarshal(tr.sent[0])
	if err != nil {
		t.Fatalf("reply unmarshal: %v", err)
	}
	if reply.Type() != Offer {
		t.Errorf("reply type = %d, want Offer", reply.Type())
	}
	if !reply.YIAddr.Equal(net.ParseIP("192.168.0.10")) {
		t.Errorf("offered IP = %v, want 192.168.0.10", reply.YIAddr)
	}
}

func TestDhcprelayValidatesAndGates(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	stdio := command.IO{In: bytes.NewReader(nil), Out: out, Err: &bytes.Buffer{}}
	err := NewDhcprelay().Run(context.Background(), stdio, []string{"eth0", "192.168.0.1"})
	if err == nil {
		t.Fatal("dhcprelay should fail with a documented capability error")
	}
	if !strings.Contains(err.Error(), "relay plan") {
		t.Errorf("error should describe the plan: %v", err)
	}
	// Bad server address is rejected before the capability gate.
	if err := NewDhcprelay().Run(context.Background(), stdio, []string{"eth0", "notanip"}); err == nil {
		t.Error("expected validation error for bad server")
	}
}

func TestUdhcpdRequiresForeground(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewUdhcpd().Run(context.Background(), stdio, []string{"conf"}); err == nil {
		t.Error("expected error without -f")
	}
}
