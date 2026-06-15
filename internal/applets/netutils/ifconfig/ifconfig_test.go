package ifconfig

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

func links() []ipcmd.Link {
	return []ipcmd.Link{
		{Index: 1, Name: "lo", Flags: []string{"LOOPBACK", "UP"}, MTU: 65536, State: "UNKNOWN",
			Addrs: []ipcmd.Addr{{Family: "inet", CIDR: "127.0.0.1/8", Scope: "host"}}},
		{Index: 2, Name: "eth0", Flags: []string{"BROADCAST", "UP"}, MTU: 1500, MAC: "52:54:00:12:34:56",
			State: "UP", Addrs: []ipcmd.Addr{{Family: "inet", CIDR: "192.168.1.10/24", Scope: "global"}}},
		{Index: 3, Name: "down0", Flags: []string{"BROADCAST"}, MTU: 1500, State: "DOWN"},
	}
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestShowActive(t *testing.T) {
	defer SetSource(links())()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "eth0: flags=<BROADCAST,UP>  mtu 1500") {
		t.Errorf("missing eth0: %s", out)
	}
	if !strings.Contains(out, "inet 192.168.1.10  netmask 255.255.255.0") {
		t.Errorf("missing inet line: %s", out)
	}
	if !strings.Contains(out, "ether 52:54:00:12:34:56") {
		t.Errorf("missing ether line: %s", out)
	}
	// down0 is not UP, so without -a it must be hidden.
	if strings.Contains(out, "down0") {
		t.Errorf("down interface should be hidden without -a: %s", out)
	}
}

func TestShowAll(t *testing.T) {
	defer SetSource(links())()
	out, _, err := run(t, "-a")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "down0") {
		t.Errorf("-a should include down interfaces: %s", out)
	}
}

func TestShowOne(t *testing.T) {
	defer SetSource(links())()
	out, _, err := run(t, "eth0")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "eth0") || strings.Contains(out, "lo:") {
		t.Errorf("expected only eth0: %s", out)
	}
}

func TestUnknownInterface(t *testing.T) {
	defer SetSource(links())()
	if _, _, err := run(t, "nope0"); err == nil {
		t.Error("expected error for unknown interface")
	}
}

func TestConfigurationRejected(t *testing.T) {
	defer SetSource(links())()
	if _, _, err := run(t, "eth0", "192.168.1.5"); err == nil {
		t.Error("expected error for configuration attempt")
	}
}

func TestNetmaskOf(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"10.0.0.1/8":   "255.0.0.0",
		"10.0.0.1/24":  "255.255.255.0",
		"10.0.0.1/25":  "255.255.255.128",
		"10.0.0.1":     "",
		"10.0.0.1/bad": "",
	}
	for in, want := range cases {
		if got := netmaskOf(in); got != want {
			t.Errorf("netmaskOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "ifconfig" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
