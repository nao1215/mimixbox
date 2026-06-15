package ifupdown

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

const sample = `# example interfaces file
auto lo eth0

iface lo inet loopback

iface eth0 inet static
    address 192.168.1.10
    netmask 255.255.255.0
    pre-up echo preparing eth0
    up echo eth0 up
    down echo eth0 down
    post-down echo eth0 cleaned
`

func TestParseConfig(t *testing.T) {
	t.Parallel()
	cfg, err := ParseConfig(strings.NewReader(sample))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}
	if len(cfg.Auto) != 2 || cfg.Auto[1] != "eth0" {
		t.Errorf("auto = %v", cfg.Auto)
	}
	eth0 := cfg.Ifaces["eth0"]
	if eth0 == nil || eth0.Method != "static" {
		t.Fatalf("eth0 = %+v", eth0)
	}
	if eth0.Options["address"] != "192.168.1.10" {
		t.Errorf("address = %q", eth0.Options["address"])
	}
	if len(eth0.PreUp) != 1 || eth0.PreUp[0] != "echo preparing eth0" {
		t.Errorf("pre-up = %v", eth0.PreUp)
	}
}

func TestParseConfigError(t *testing.T) {
	t.Parallel()
	if _, err := ParseConfig(strings.NewReader("address 1.2.3.4")); err == nil {
		t.Error("expected error for option outside stanza")
	}
	if _, err := ParseConfig(strings.NewReader("iface eth0 inet")); err == nil {
		t.Error("expected error for incomplete iface")
	}
}

func TestPlanUpDown(t *testing.T) {
	t.Parallel()
	cfg, _ := ParseConfig(strings.NewReader(sample))
	up, err := Plan(cfg, "eth0", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(up) != 2 || up[0] != "echo preparing eth0" || up[1] != "echo eth0 up" {
		t.Errorf("up plan = %v", up)
	}
	down, _ := Plan(cfg, "eth0", false)
	if len(down) != 2 || down[0] != "echo eth0 down" || down[1] != "echo eth0 cleaned" {
		t.Errorf("down plan = %v", down)
	}
	if _, err := Plan(cfg, "nope", true); err == nil {
		t.Error("expected error for unknown interface")
	}
}

func TestRunHooksWithInjectedRunner(t *testing.T) {
	t.Parallel()
	var ran []string
	runner := func(_ context.Context, cmd string) error {
		ran = append(ran, cmd)
		return nil
	}
	cmds := []string{"echo a", "", "echo b"}
	if err := RunHooks(context.Background(), cmds, runner); err != nil {
		t.Fatal(err)
	}
	if len(ran) != 2 || ran[0] != "echo a" || ran[1] != "echo b" {
		t.Errorf("ran = %v", ran)
	}
}

func TestIfupNoActPrintsPlan(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "interfaces")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	stdio := command.IO{In: bytes.NewReader(nil), Out: out, Err: &bytes.Buffer{}}
	if err := NewIfup().Run(context.Background(), stdio, []string{"-n", "-i", path, "eth0"}); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "echo preparing eth0") || !strings.Contains(got, "echo eth0 up") {
		t.Errorf("no-act output missing hooks:\n%s", got)
	}
}

func TestIfupRunsHooksThenGates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "interfaces")
	if err := os.WriteFile(path, []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	stdio := command.IO{In: bytes.NewReader(nil), Out: out, Err: &bytes.Buffer{}}
	// Without -n: hooks run (echo writes to out) then the state change is gated.
	err := NewIfup().Run(context.Background(), stdio, []string{"-i", path, "eth0"})
	if err == nil || !strings.Contains(err.Error(), "capability-gated") {
		t.Fatalf("expected capability error, got %v", err)
	}
	if !strings.Contains(out.String(), "eth0 up") {
		t.Errorf("hooks did not run; out=%q", out.String())
	}
}

func TestIfplugdCapabilityGated(t *testing.T) {
	t.Parallel()
	stdio := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := NewIfplugd().Run(context.Background(), stdio, []string{"-i", "eth0"}); err == nil {
		t.Error("ifplugd should be capability-gated")
	}
	if err := NewIfplugd().Run(context.Background(), stdio, nil); err == nil {
		t.Error("ifplugd should require -i")
	}
}
