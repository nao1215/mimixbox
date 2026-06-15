package pscan

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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

func stub(t *testing.T, open map[int]bool) {
	t.Helper()
	orig := probe
	probe = func(_ string, port int, _ time.Duration) bool { return open[port] }
	t.Cleanup(func() { probe = orig })
}

func TestReportsOpenPorts(t *testing.T) {
	stub(t, map[int]bool{22: true, 80: true})
	out, _, err := run(t, "-p", "20", "-P", "100", "127.0.0.1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "22: open") || !strings.Contains(out, "80: open") {
		t.Errorf("missing open ports: %s", out)
	}
	if strings.Contains(out, "21: open") {
		t.Errorf("closed port reported: %s", out)
	}
}

func TestNoOpenPorts(t *testing.T) {
	stub(t, map[int]bool{})
	out, _, err := run(t, "-p", "1", "-P", "10", "127.0.0.1")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no output, got %q", out)
	}
}

func TestBadRange(t *testing.T) {
	stub(t, map[int]bool{})
	if _, _, err := run(t, "-p", "100", "-P", "10", "127.0.0.1"); err == nil {
		t.Error("expected error for inverted range")
	}
	if _, _, err := run(t, "-p", "0", "127.0.0.1"); err == nil {
		t.Error("expected error for port 0")
	}
}

func TestMissingHost(t *testing.T) {
	stub(t, map[int]bool{})
	if _, _, err := run(t); err == nil {
		t.Error("expected error with no host")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "pscan" || c.Synopsis() == "" {
		t.Errorf("Name/Synopsis: %q / %q", c.Name(), c.Synopsis())
	}
}
