package linkadmin

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, cmd *Command, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cmd.Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestTCInspect(t *testing.T) {
	out, _, err := run(t, NewTC(), "qdisc", "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "no entries") {
		t.Errorf("expected inspect notice: %s", out)
	}
}

func TestTCMutatingDeferred(t *testing.T) {
	_, _, err := run(t, NewTC(), "qdisc", "add", "dev", "eth0", "root")
	if err == nil {
		t.Fatal("expected error for mutating tc subcommand")
	}
	if !strings.Contains(err.Error(), "deferred") {
		t.Errorf("err = %v", err)
	}
}

func TestIPTunnelInspect(t *testing.T) {
	out, _, err := run(t, NewIPTunnel(), "show")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "no entries") {
		t.Errorf("expected inspect notice: %s", out)
	}
}

func TestIPTunnelMutatingDeferred(t *testing.T) {
	if _, _, err := run(t, NewIPTunnel(), "add", "tun0"); err == nil {
		t.Error("expected error for mutating iptunnel subcommand")
	}
}

func TestNameifDeferred(t *testing.T) {
	_, _, err := run(t, NewNameif(), "eth0", "00:11:22:33:44:55")
	if err == nil {
		t.Fatal("expected capability error")
	}
	if !strings.Contains(err.Error(), "deferred") {
		t.Errorf("err = %v", err)
	}
}

func TestNameifBadArgs(t *testing.T) {
	if _, _, err := run(t, NewNameif(), "eth0"); err == nil {
		t.Error("expected usage error")
	}
}

func TestSlattachDeferred(t *testing.T) {
	_, _, err := run(t, NewSlattach(), "-p", "slip", "/dev/ttyS0")
	if err == nil {
		t.Fatal("expected capability error")
	}
	if !strings.Contains(err.Error(), "deferred") {
		t.Errorf("err = %v", err)
	}
}

func TestSlattachBadArgs(t *testing.T) {
	if _, _, err := run(t, NewSlattach()); err == nil {
		t.Error("expected usage error")
	}
}

func TestNamesAndSynopses(t *testing.T) {
	t.Parallel()
	cmds := map[string]*Command{
		"tc":       NewTC(),
		"iptunnel": NewIPTunnel(),
		"nameif":   NewNameif(),
		"slattach": NewSlattach(),
	}
	for want, c := range cmds {
		if c.Name() != want {
			t.Errorf("Name() = %q, want %q", c.Name(), want)
		}
		if c.Synopsis() == "" {
			t.Errorf("%s Synopsis() empty", want)
		}
	}
}
