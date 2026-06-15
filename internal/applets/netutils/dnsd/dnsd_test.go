package dnsd

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestParseHosts(t *testing.T) {
	t.Parallel()
	zone, err := ParseHosts(strings.NewReader("# c\n10.0.0.1 alpha\n10.0.0.2 Beta.\n"))
	if err != nil {
		t.Fatalf("ParseHosts error: %v", err)
	}
	if zone["alpha"].String() != "10.0.0.1" {
		t.Errorf("alpha = %v", zone["alpha"])
	}
	if zone["beta"].String() != "10.0.0.2" {
		t.Errorf("beta = %v", zone["beta"])
	}
}

func TestParseHostsError(t *testing.T) {
	t.Parallel()
	if _, err := ParseHosts(strings.NewReader("notanip name")); err == nil {
		t.Error("expected error for bad IP")
	}
	if _, err := ParseHosts(strings.NewReader("onlyonefield")); err == nil {
		t.Error("expected error for missing name")
	}
}

// buildQuery encodes a minimal A/IN query for name with id.
func buildQuery(id uint16, name string) []byte {
	q := make([]byte, 12)
	binary.BigEndian.PutUint16(q[0:2], id)
	q[2] = 0x01 // RD
	binary.BigEndian.PutUint16(q[4:6], 1)
	for _, label := range strings.Split(name, ".") {
		q = append(q, byte(len(label)))
		q = append(q, label...)
	}
	q = append(q, 0)
	q = append(q, 0x00, 0x01, 0x00, 0x01) // QTYPE A, QCLASS IN
	return q
}

func TestBuildResponseKnownName(t *testing.T) {
	t.Parallel()
	zone := map[string]net.IP{"alpha": net.ParseIP("192.0.2.5").To4()}
	resp, err := BuildResponse(buildQuery(0x1234, "alpha"), zone)
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}
	if binary.BigEndian.Uint16(resp[0:2]) != 0x1234 {
		t.Errorf("id mismatch")
	}
	if binary.BigEndian.Uint16(resp[6:8]) != 1 {
		t.Fatalf("ANCOUNT = %d, want 1", binary.BigEndian.Uint16(resp[6:8]))
	}
	ip := resp[len(resp)-4:]
	if !bytes.Equal(ip, []byte{192, 0, 2, 5}) {
		t.Errorf("answer IP = %v", ip)
	}
}

func TestBuildResponseUnknownName(t *testing.T) {
	t.Parallel()
	resp, err := BuildResponse(buildQuery(1, "missing"), map[string]net.IP{})
	if err != nil {
		t.Fatalf("BuildResponse error: %v", err)
	}
	if resp[3]&0x0f != 3 {
		t.Errorf("rcode = %d, want 3 (NXDOMAIN)", resp[3]&0x0f)
	}
}

func TestServeOverLoopback(t *testing.T) {
	t.Parallel()
	pc, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Skipf("loopback UDP unavailable: %v", err)
	}
	zone := map[string]net.IP{"host.local": net.ParseIP("127.0.0.99").To4()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = Serve(ctx, pc, zone)
	}()

	client, err := net.DialUDP("udp", nil, pc.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = client.Close() }()
	_, _ = client.Write(buildQuery(7, "host.local"))
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 512)
	n, _, err := client.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(buf[n-4:n], []byte{127, 0, 0, 99}) {
		t.Errorf("answer = %v", buf[n-4:n])
	}

	cancel()
	wg.Wait()
}

func TestRunValidation(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), stdio, []string{"-H", "x"}); err == nil {
		t.Error("expected error without -f")
	}
	if err := New().Run(context.Background(), stdio, []string{"-f"}); err == nil {
		t.Error("expected error without -H")
	}
}
