package ipcalc

import (
	"bytes"
	"context"
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

func TestComputeValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		addr      string
		mask      string // NETMASK operand
		flag      string // -m
		wantNet   string
		wantBcast string
		wantMask  string
		wantPfx   int
		wantMin   string
		wantMax   string
		wantHosts uint64
	}{
		{
			name: "slash24", addr: "192.168.10.7/24",
			wantNet: "192.168.10.0", wantBcast: "192.168.10.255", wantMask: "255.255.255.0",
			wantPfx: 24, wantMin: "192.168.10.1", wantMax: "192.168.10.254", wantHosts: 254,
		},
		{
			name: "slash20", addr: "172.16.5.9/20",
			wantNet: "172.16.0.0", wantBcast: "172.16.15.255", wantMask: "255.255.240.0",
			wantPfx: 20, wantMin: "172.16.0.1", wantMax: "172.16.15.254", wantHosts: 4094,
		},
		{
			name: "dotted mask operand", addr: "10.0.0.1", mask: "255.255.255.0",
			wantNet: "10.0.0.0", wantBcast: "10.0.0.255", wantMask: "255.255.255.0",
			wantPfx: 24, wantMin: "10.0.0.1", wantMax: "10.0.0.254", wantHosts: 254,
		},
		{
			name: "slash25", addr: "192.168.1.130/25",
			wantNet: "192.168.1.128", wantBcast: "192.168.1.255", wantMask: "255.255.255.128",
			wantPfx: 25, wantMin: "192.168.1.129", wantMax: "192.168.1.254", wantHosts: 126,
		},
		{
			name: "slash31 point-to-point", addr: "10.1.1.0/31",
			wantNet: "10.1.1.0", wantBcast: "10.1.1.1", wantMask: "255.255.255.254",
			wantPfx: 31, wantMin: "10.1.1.0", wantMax: "10.1.1.1", wantHosts: 2,
		},
		{
			name: "slash32 host", addr: "8.8.8.8/32",
			wantNet: "8.8.8.8", wantBcast: "8.8.8.8", wantMask: "255.255.255.255",
			wantPfx: 32, wantMin: "8.8.8.8", wantMax: "8.8.8.8", wantHosts: 1,
		},
		{
			name: "slash8", addr: "10.20.30.40/8",
			wantNet: "10.0.0.0", wantBcast: "10.255.255.255", wantMask: "255.0.0.0",
			wantPfx: 8, wantMin: "10.0.0.1", wantMax: "10.255.255.254", wantHosts: 16777214,
		},
		{
			name: "slash0 whole space", addr: "1.2.3.4/0",
			wantNet: "0.0.0.0", wantBcast: "255.255.255.255", wantMask: "0.0.0.0",
			wantPfx: 0, wantMin: "0.0.0.1", wantMax: "255.255.255.254", wantHosts: 4294967294,
		},
		{
			name: "mask via -m prefix", addr: "192.168.0.5", flag: "26",
			wantNet: "192.168.0.0", wantBcast: "192.168.0.63", wantMask: "255.255.255.192",
			wantPfx: 26, wantMin: "192.168.0.1", wantMax: "192.168.0.62", wantHosts: 62,
		},
		{
			name: "classful default class C", addr: "192.168.1.1",
			wantNet: "192.168.1.0", wantBcast: "192.168.1.255", wantMask: "255.255.255.0",
			wantPfx: 24, wantMin: "192.168.1.1", wantMax: "192.168.1.254", wantHosts: 254,
		},
		{
			name: "classful default class A", addr: "10.5.6.7",
			wantNet: "10.0.0.0", wantBcast: "10.255.255.255", wantMask: "255.0.0.0",
			wantPfx: 8, wantMin: "10.0.0.1", wantMax: "10.255.255.254", wantHosts: 16777214,
		},
		{
			name: "classful default class B", addr: "172.16.5.4",
			wantNet: "172.16.0.0", wantBcast: "172.16.255.255", wantMask: "255.255.0.0",
			wantPfx: 16, wantMin: "172.16.0.1", wantMax: "172.16.255.254", wantHosts: 65534,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := compute(tt.addr, tt.mask, tt.flag)
			if err != nil {
				t.Fatalf("compute(%q,%q,%q) error = %v", tt.addr, tt.mask, tt.flag, err)
			}
			if res.network != tt.wantNet {
				t.Errorf("network = %q, want %q", res.network, tt.wantNet)
			}
			if res.broadcast != tt.wantBcast {
				t.Errorf("broadcast = %q, want %q", res.broadcast, tt.wantBcast)
			}
			if res.netmask != tt.wantMask {
				t.Errorf("netmask = %q, want %q", res.netmask, tt.wantMask)
			}
			if res.prefix != tt.wantPfx {
				t.Errorf("prefix = %d, want %d", res.prefix, tt.wantPfx)
			}
			if res.hostMin != tt.wantMin {
				t.Errorf("hostMin = %q, want %q", res.hostMin, tt.wantMin)
			}
			if res.hostMax != tt.wantMax {
				t.Errorf("hostMax = %q, want %q", res.hostMax, tt.wantMax)
			}
			if res.hostCount != tt.wantHosts {
				t.Errorf("hostCount = %d, want %d", res.hostCount, tt.wantHosts)
			}
		})
	}
}

func TestComputeInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, addr, mask, flag string
	}{
		{"bad address", "999.1.1.1", "", ""},
		{"ipv6 rejected", "2001:db8::1/64", "", ""},
		{"prefix too big", "10.0.0.1/33", "", ""},
		{"negative prefix", "10.0.0.1/-1", "", ""},
		{"non-numeric prefix", "10.0.0.1/ab", "", ""},
		{"bad dotted mask", "10.0.0.1", "255.0.255.0", ""},
		{"bad mask operand", "10.0.0.1", "not-a-mask", ""},
		{"empty address", "", "", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := compute(tt.addr, tt.mask, tt.flag); err == nil {
				t.Errorf("compute(%q,%q,%q) expected error", tt.addr, tt.mask, tt.flag)
			}
		})
	}
}

func TestRunTable(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "192.168.10.7/24")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{
		"Address:   192.168.10.7",
		"Netmask:   255.255.255.0 = 24",
		"Network:   192.168.10.0/24",
		"Broadcast: 192.168.10.255",
		"HostMin:   192.168.10.1",
		"HostMax:   192.168.10.254",
		"Hosts:     254",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestRunSelective(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-b", "-n", "172.16.5.9/20")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "NETWORK=172.16.0.0\n") {
		t.Errorf("missing NETWORK line: %q", out)
	}
	if !strings.Contains(out, "BROADCAST=172.16.15.255\n") {
		t.Errorf("missing BROADCAST line: %q", out)
	}
	if strings.Contains(out, "Address:") {
		t.Errorf("selective mode should not print the table: %q", out)
	}
}

func TestRunPrefixOnly(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "-p", "192.168.1.1", "255.255.255.128")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) != "PREFIX=25" {
		t.Errorf("out = %q, want PREFIX=25", out)
	}
}

func TestRunHostrange(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--hostrange", "10.0.0.0/30")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{"HOSTMIN=10.0.0.1\n", "HOSTMAX=10.0.0.2\n", "HOSTS=2\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t); err == nil {
		t.Error("expected error with no operands")
	}
	if _, _, err := run(t, "a", "b", "c"); err == nil {
		t.Error("expected error with too many operands")
	}
	if _, _, err := run(t, "bad-ip/24"); err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "ipcalc" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
