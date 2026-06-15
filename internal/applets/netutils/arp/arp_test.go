package arp

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/netutils/ipcmd"
	"github.com/nao1215/mimixbox/internal/command"
)

func neighbours() []ipcmd.Neighbour {
	return []ipcmd.Neighbour{
		{IP: "192.168.1.1", Dev: "eth0", MAC: "52:54:00:aa:bb:cc", State: "REACHABLE"},
		{IP: "192.168.1.99", Dev: "eth0", MAC: "", State: "FAILED"},
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
	defer SetSource(neighbours())()
	out, _, err := run(t, "-n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "Address") || !strings.Contains(out, "HWaddress") {
		t.Errorf("missing header: %s", out)
	}
	if !strings.Contains(out, "192.168.1.1") || !strings.Contains(out, "52:54:00:aa:bb:cc") {
		t.Errorf("missing reachable entry: %s", out)
	}
	if !strings.Contains(out, "(incomplete)") {
		t.Errorf("incomplete entry should be marked: %s", out)
	}
}

func TestSetRejected(t *testing.T) {
	defer SetSource(neighbours())()
	if _, _, err := run(t, "-s", "1.2.3.4", "aa:bb:cc:dd:ee:ff"); err == nil {
		t.Error("expected error for arp -s")
	}
	if _, _, err := run(t, "-d", "1.2.3.4"); err == nil {
		t.Error("expected error for arp -d")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "arp" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
