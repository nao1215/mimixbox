package whris

import (
	"bytes"
	"context"
	"errors"
	"net"
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

func stub(t *testing.T, ips []net.IP, asn, owner string, lookErr error) {
	t.Helper()
	origR, origA := resolver, asnLookup
	resolver = func(string) ([]net.IP, error) { return ips, nil }
	asnLookup = func(string) (string, string, error) { return asn, owner, lookErr }
	t.Cleanup(func() { resolver, asnLookup = origR, origA })
}

func TestReportsInfo(t *testing.T) {
	stub(t, []net.IP{net.ParseIP("93.184.216.34")}, "15133", "EDGECAST", nil)
	out, _, err := run(t, "example.com")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "93.184.216.34\tAS15133\tEDGECAST") {
		t.Errorf("out = %q", out)
	}
}

func TestLookupFailureIsTolerated(t *testing.T) {
	stub(t, []net.IP{net.ParseIP("1.2.3.4")}, "", "", errors.New("boom"))
	out, _, err := run(t, "example.com")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "1.2.3.4\tAS?\tlookup failed") {
		t.Errorf("out = %q", out)
	}
}

func TestSkipsIPv6(t *testing.T) {
	t.Parallel()
	infos := collect([]net.IP{net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")})
	if len(infos) != 0 {
		t.Errorf("IPv6 addresses should be skipped, got %v", infos)
	}
}

func TestResolveFailure(t *testing.T) {
	origR := resolver
	resolver = func(string) ([]net.IP, error) { return nil, errors.New("nxdomain") }
	t.Cleanup(func() { resolver = origR })

	_, _, err := run(t, "example.com")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot resolve") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingDomain(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exactly one DOMAIN") {
		t.Errorf("err = %v", err)
	}
}

func TestParseCymru(t *testing.T) {
	t.Parallel()
	resp := "AS      | IP               | BGP Prefix          | CC | Registry | Allocated  | AS Name\n" +
		"15169   | 8.8.8.8          | 8.8.8.0/24          | US | arin     | 1992-12-01 | GOOGLE, US\n"
	asn, owner, err := parseCymru(strings.NewReader(resp))
	if err != nil {
		t.Fatalf("parseCymru error = %v", err)
	}
	if asn != "15169" {
		t.Errorf("asn = %q, want 15169", asn)
	}
	if owner != "GOOGLE, US" {
		t.Errorf("owner = %q", owner)
	}
}

func TestParseCymruNoData(t *testing.T) {
	t.Parallel()
	_, _, err := parseCymru(strings.NewReader("AS | IP | name\n"))
	if err == nil {
		t.Error("expected error for header-only response")
	}
}

func TestCymruLookupAgainstLocalServer(t *testing.T) {
	// Not parallel: this test mutates the shared cymruServer global, so running it
	// alongside another cymruServer-mutating test would race and could reach the
	// real Cymru server.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("loopback TCP/UDP listen unavailable: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 128)
		_, _ = conn.Read(buf)
		_, _ = conn.Write([]byte(
			"AS | IP | BGP Prefix | CC | Registry | Allocated | AS Name\n" +
				"13335 | 1.1.1.1 | 1.1.1.0/24 | US | arin | 2010-07-14 | CLOUDFLARENET, US\n"))
	}()

	orig := cymruServer
	cymruServer = ln.Addr().String()
	t.Cleanup(func() { cymruServer = orig })

	asn, owner, err := cymruLookup("1.1.1.1")
	if err != nil {
		t.Fatalf("cymruLookup error = %v", err)
	}
	if asn != "13335" || owner != "CLOUDFLARENET, US" {
		t.Errorf("got AS%s %q", asn, owner)
	}
}

func TestCymruLookupDialError(t *testing.T) {
	// Not parallel: mutates the shared cymruServer global (see above).
	orig := cymruServer
	cymruServer = "127.0.0.1:1" // closed port
	t.Cleanup(func() { cymruServer = orig })
	if _, _, err := cymruLookup("8.8.8.8"); err == nil {
		t.Error("expected a dial error")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "whris" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
