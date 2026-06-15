package etherwake

import (
	"bytes"
	"context"
	"errors"
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

func stub(t *testing.T) *[]byte {
	t.Helper()
	var captured []byte
	orig := send
	send = func(_ string, payload []byte) error { captured = payload; return nil }
	t.Cleanup(func() { send = orig })
	return &captured
}

func TestMagicPacket(t *testing.T) {
	t.Parallel()
	p, err := magicPacket("11:22:33:44:55:66")
	if err != nil {
		t.Fatalf("magicPacket error = %v", err)
	}
	if len(p) != 102 {
		t.Fatalf("len = %d, want 102", len(p))
	}
	for i := 0; i < 6; i++ {
		if p[i] != 0xFF {
			t.Fatalf("byte %d = %#x, want 0xFF", i, p[i])
		}
	}
	mac := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
	for rep := 0; rep < 16; rep++ {
		off := 6 + rep*6
		if !bytes.Equal(p[off:off+6], mac) {
			t.Fatalf("repetition %d = %x, want %x", rep, p[off:off+6], mac)
		}
	}
}

func TestMagicPacketInvalidMAC(t *testing.T) {
	t.Parallel()
	if _, err := magicPacket("not-a-mac"); err == nil {
		t.Error("expected error for invalid MAC")
	}
}

func TestRunSends(t *testing.T) {
	captured := stub(t)
	out, _, err := run(t, "11:22:33:44:55:66")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if len(*captured) != 102 {
		t.Errorf("captured packet len = %d, want 102", len(*captured))
	}
	if !strings.Contains(out, "Sent magic packet to 11:22:33:44:55:66") {
		t.Errorf("missing confirmation: %s", out)
	}
}

func TestSendError(t *testing.T) {
	orig := send
	send = func(string, []byte) error { return errors.New("network down") }
	t.Cleanup(func() { send = orig })
	if _, _, err := run(t, "11:22:33:44:55:66"); err == nil {
		t.Error("expected error when send fails")
	}
}

func TestRawSocketDeferred(t *testing.T) {
	stub(t)
	if _, _, err := run(t, "-i", "eth0", "11:22:33:44:55:66"); err == nil {
		t.Error("expected error for -i (raw socket deferred)")
	}
}

func TestBadArgs(t *testing.T) {
	stub(t)
	if _, _, err := run(t); err == nil {
		t.Error("expected error with no MAC")
	}
	if _, _, err := run(t, "a:b", "c:d"); err == nil {
		t.Error("expected error with two operands")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "ether-wake" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
