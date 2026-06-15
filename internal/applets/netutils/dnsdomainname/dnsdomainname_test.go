package dnsdomainname

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

func stub(t *testing.T, host string, hostErr error, addrs []string, addrErr error, names []string, nameErr error) {
	t.Helper()
	oh, oa, on := hostnameFn, lookupHostFn, lookupAddrFn
	hostnameFn = func() (string, error) { return host, hostErr }
	lookupHostFn = func(string) ([]string, error) { return addrs, addrErr }
	lookupAddrFn = func(string) ([]string, error) { return names, nameErr }
	t.Cleanup(func() { hostnameFn, lookupHostFn, lookupAddrFn = oh, oa, on })
}

func TestDomainFromReverseLookup(t *testing.T) {
	stub(t, "node1", nil, []string{"10.0.0.1"}, nil, []string{"node1.example.com."}, nil)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) != "example.com" {
		t.Errorf("out = %q, want example.com", out)
	}
}

func TestDomainFromHostnameWithDot(t *testing.T) {
	stub(t, "web.corp.internal", nil, nil, nil, nil, nil)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) != "corp.internal" {
		t.Errorf("out = %q, want corp.internal", out)
	}
}

func TestNoDomainPrintsNothing(t *testing.T) {
	stub(t, "localhost", nil, []string{"127.0.0.1"}, nil, []string{"localhost."}, nil)
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

func TestAddressLookupFailure(t *testing.T) {
	stub(t, "node1", nil, nil, errors.New("no addr"), nil, nil)
	if _, _, err := run(t); err == nil {
		t.Error("expected error when address lookup fails")
	}
}

func TestReverseLookupFailure(t *testing.T) {
	stub(t, "node1", nil, []string{"10.0.0.1"}, nil, nil, errors.New("nxdomain"))
	if _, _, err := run(t); err == nil {
		t.Error("expected error when reverse lookup fails")
	}
}

func TestRejectsOperands(t *testing.T) {
	stub(t, "host.example.com", nil, nil, nil, nil, nil)
	if _, _, err := run(t, "extra"); err == nil {
		t.Error("expected error for unexpected operand")
	}
}

func TestDomainOf(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"a.b.c":     "b.c",
		"host.dom":  "dom",
		"single":    "",
		"x.y.z.tld": "y.z.tld",
	}
	for in, want := range cases {
		if got := domainOf(in); got != want {
			t.Errorf("domainOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "dnsdomainname" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
