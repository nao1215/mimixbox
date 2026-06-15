package route

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

func routes() []ipcmd.Route {
	return []ipcmd.Route{
		{Dest: "default", Via: "192.168.1.1", Dev: "eth0"},
		{Dest: "192.168.1.0/24", Dev: "eth0", Src: "192.168.1.10"},
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

func TestShowTable(t *testing.T) {
	defer SetSource(routes())()
	out, _, err := run(t, "-n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Destination") || !strings.Contains(out, "Gateway") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "0.0.0.0") || !strings.Contains(out, "192.168.1.1") {
		t.Errorf("missing default route: %s", out)
	}
	if !strings.Contains(out, "255.255.255.0") {
		t.Errorf("missing genmask for /24: %s", out)
	}
	// default route is a gateway, so flags must include G.
	if !strings.Contains(out, "UG") {
		t.Errorf("default route should have UG flag: %s", out)
	}
}

func TestDestAndMask(t *testing.T) {
	t.Parallel()
	d, m := destAndMask("default")
	if d != "0.0.0.0" || m != "0.0.0.0" {
		t.Errorf("default = %q/%q", d, m)
	}
	d, m = destAndMask("10.0.0.0/8")
	if d != "10.0.0.0" || m != "255.0.0.0" {
		t.Errorf("10.0.0.0/8 = %q/%q", d, m)
	}
}

func TestAddRejected(t *testing.T) {
	defer SetSource(routes())()
	if _, _, err := run(t, "add", "default"); err == nil {
		t.Error("expected error for route add")
	}
	if _, _, err := run(t, "del", "default"); err == nil {
		t.Error("expected error for route del")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "route" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
