package nslookup

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// fakeResolver answers from in-memory maps.
type fakeResolver struct {
	hosts map[string][]string
	addrs map[string][]string
	err   error
}

func (f fakeResolver) LookupHost(_ context.Context, host string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.hosts[host], nil
}

func (f fakeResolver) LookupAddr(_ context.Context, addr string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.addrs[addr], nil
}

func stub(t *testing.T, r Resolver) {
	t.Helper()
	orig := newResolver
	newResolver = func(string) Resolver { return r }
	t.Cleanup(func() { newResolver = orig })
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestForwardLookup(t *testing.T) {
	stub(t, fakeResolver{hosts: map[string][]string{"example.test": {"192.0.2.10", "192.0.2.11"}}})
	out, _, err := run(t, "example.test")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Name:\texample.test") {
		t.Errorf("missing name: %s", out)
	}
	if !strings.Contains(out, "Address: 192.0.2.10") || !strings.Contains(out, "Address: 192.0.2.11") {
		t.Errorf("missing addresses: %s", out)
	}
}

func TestForwardWithServer(t *testing.T) {
	stub(t, fakeResolver{hosts: map[string][]string{"example.test": {"192.0.2.10"}}})
	out, _, err := run(t, "example.test", "127.0.0.1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Server:\t\t127.0.0.1") {
		t.Errorf("server header missing: %s", out)
	}
}

func TestReverseLookup(t *testing.T) {
	stub(t, fakeResolver{addrs: map[string][]string{"192.0.2.10": {"host.example.test."}}})
	out, _, err := run(t, "192.0.2.10")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "192.0.2.10\tname = host.example.test") {
		t.Errorf("missing PTR: %s", out)
	}
}

func TestLookupFailure(t *testing.T) {
	stub(t, fakeResolver{err: errors.New("nxdomain")})
	if _, _, err := run(t, "nope.test"); err == nil {
		t.Error("expected error on resolver failure")
	}
}

func TestNoRecords(t *testing.T) {
	stub(t, fakeResolver{hosts: map[string][]string{}})
	if _, _, err := run(t, "empty.test"); err == nil {
		t.Error("expected error when no records returned")
	}
}

func TestBadArgs(t *testing.T) {
	stub(t, fakeResolver{})
	if _, _, err := run(t); err == nil {
		t.Error("expected error with no operands")
	}
	if _, _, err := run(t, "a", "b", "c"); err == nil {
		t.Error("expected error with too many operands")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "nslookup" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
