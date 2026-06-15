package netstat

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func sockets() []Socket {
	return []Socket{
		{Proto: "tcp", Local: "0.0.0.0:22", Foreign: "0.0.0.0:*", State: "LISTEN"},
		{Proto: "tcp", Local: "192.168.1.10:54321", Foreign: "93.184.216.34:443", State: "ESTABLISHED"},
		{Proto: "udp", Local: "0.0.0.0:68", Foreign: "0.0.0.0:*", State: ""},
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

func TestShowAll(t *testing.T) {
	defer SetSource(sockets())()
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, want := range []string{"0.0.0.0:22", "93.184.216.34:443", "0.0.0.0:68"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q: %s", want, out)
		}
	}
}

func TestListeningOnly(t *testing.T) {
	defer SetSource(sockets())()
	out, _, err := run(t, "-tln")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "0.0.0.0:22") {
		t.Errorf("listening socket missing: %s", out)
	}
	if strings.Contains(out, "ESTABLISHED") {
		t.Errorf("-l should hide established connections: %s", out)
	}
	if strings.Contains(out, ":68") {
		t.Errorf("-t should hide UDP: %s", out)
	}
}

func TestUDPOnly(t *testing.T) {
	defer SetSource(sockets())()
	out, _, err := run(t, "-u")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, ":68") {
		t.Errorf("expected UDP socket: %s", out)
	}
	if strings.Contains(out, ":22") {
		t.Errorf("-u should hide TCP: %s", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "netstat" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
