package hostid

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestHostIDFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "hostid")
	// Little-endian 0x078bbefa -> bytes fa be 8b 07.
	if err := os.WriteFile(path, []byte{0xfa, 0xbe, 0x8b, 0x07}, 0o600); err != nil {
		t.Fatal(err)
	}

	got, ok := hostIDFromFile(path)
	if !ok {
		t.Fatal("hostIDFromFile ok = false, want true")
	}
	if got != 0x078bbefa {
		t.Errorf("hostIDFromFile = %08x, want 078bbefa", got)
	}
}

func TestHostIDFromFileMissingOrShort(t *testing.T) {
	t.Parallel()

	if _, ok := hostIDFromFile(filepath.Join(t.TempDir(), "absent")); ok {
		t.Error("missing file: ok = true, want false")
	}

	short := filepath.Join(t.TempDir(), "short")
	if err := os.WriteFile(short, []byte{0x01, 0x02}, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := hostIDFromFile(short); ok {
		t.Error("short file: ok = true, want false")
	}
}

func TestHostIDFromHostname(t *testing.T) {
	t.Parallel()

	hostname := func() (string, error) { return "host", nil }
	lookup := func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}

	// 127.0.0.1 little-endian = 0x0100007f; swapping halves -> 0x007f0100.
	if got := hostIDFromHostname(hostname, lookup); got != 0x007f0100 {
		t.Errorf("hostIDFromHostname = %08x, want 007f0100", got)
	}
}

func TestHostIDFromHostnameErrors(t *testing.T) {
	t.Parallel()

	noName := func() (string, error) { return "", os.ErrInvalid }
	okLookup := func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("10.0.0.1")}, nil }
	if got := hostIDFromHostname(noName, okLookup); got != 0 {
		t.Errorf("hostname error: got %08x, want 0", got)
	}

	okName := func() (string, error) { return "host", nil }
	noLookup := func(string) ([]net.IP, error) { return nil, os.ErrNotExist }
	if got := hostIDFromHostname(okName, noLookup); got != 0 {
		t.Errorf("lookup error: got %08x, want 0", got)
	}

	// No IPv4 address available -> 0.
	v6Only := func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("::1")}, nil }
	if got := hostIDFromHostname(okName, v6Only); got != 0 {
		t.Errorf("ipv6 only: got %08x, want 0", got)
	}
}

func TestHostIDPrefersFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hostid")
	if err := os.WriteFile(path, []byte{0x78, 0x56, 0x34, 0x12}, 0o600); err != nil {
		t.Fatal(err)
	}

	old := hostidFile
	hostidFile = path
	defer func() { hostidFile = old }()

	if got := hostID(); got != 0x12345678 {
		t.Errorf("hostID = %08x, want 12345678", got)
	}
}
