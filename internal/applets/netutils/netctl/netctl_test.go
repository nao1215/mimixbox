package netctl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestPlans(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cmd     *Command
		args    []string
		wantStr string
	}{
		{"brctl-addbr", NewBrctl(), []string{"addbr", "br0"}, "brctl addbr br0"},
		{"brctl-addif", NewBrctl(), []string{"addif", "br0", "eth0"}, "brctl addif br0 eth0"},
		{"ifenslave", NewIfenslave(), []string{"bond0", "eth0", "eth1"}, "ifenslave enslave bond0 eth0 eth1"},
		{"tunctl-create", NewTunctl(), []string{"-t", "tap0"}, "tunctl create tap0"},
		{"tunctl-delete", NewTunctl(), []string{"-d", "tap0"}, "tunctl delete tap0"},
		{"vconfig-add", NewVconfig(), []string{"add", "eth0", "100"}, "vconfig add eth0 100"},
		{"zcip", NewZcip(), []string{"eth0", "/etc/zcip.script"}, "zcip configure eth0 /etc/zcip.script"},
		{"nbd-connect", NewNbdClient(), []string{"10.0.0.1", "10809", "/dev/nbd0"}, "nbd-client connect 10.0.0.1 10809 /dev/nbd0"},
		{"nbd-disconnect", NewNbdClient(), []string{"-d", "/dev/nbd0"}, "nbd-client disconnect /dev/nbd0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := tt.cmd.plan(tt.args)
			if err != nil {
				t.Fatalf("plan error: %v", err)
			}
			if p.String() != tt.wantStr {
				t.Errorf("plan = %q, want %q", p.String(), tt.wantStr)
			}
		})
	}
}

func TestPlanValidation(t *testing.T) {
	t.Parallel()
	bad := []struct {
		cmd  *Command
		args []string
	}{
		{NewBrctl(), nil},
		{NewBrctl(), []string{"bogus"}},
		{NewBrctl(), []string{"addif", "br0"}},
		{NewIfenslave(), []string{"bond0"}},
		{NewTunctl(), nil},
		{NewVconfig(), []string{"add", "eth0", "9999"}},
		{NewVconfig(), []string{"add", "eth0"}},
		{NewZcip(), []string{"eth0"}},
		{NewNbdClient(), []string{"host", "notaport", "/dev/nbd0"}},
		{NewNbdClient(), []string{"10.0.0.1", "10809"}},
	}
	for i, tt := range bad {
		if _, err := tt.cmd.plan(tt.args); err == nil {
			t.Errorf("case %d (%s %v): expected validation error", i, tt.cmd.Name(), tt.args)
		}
	}
}

func TestRunGatesAfterValidation(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	stdio := command.IO{In: bytes.NewReader(nil), Out: out, Err: &bytes.Buffer{}}
	err := NewBrctl().Run(context.Background(), stdio, []string{"addbr", "br0"})
	if err == nil || !strings.Contains(err.Error(), "capability-gated") {
		t.Fatalf("expected capability error, got %v", err)
	}
	if !strings.Contains(out.String(), "planned action: brctl addbr br0") {
		t.Errorf("plan not reported on stdout: %q", out.String())
	}
	// Invalid args fail before the gate, without printing a plan.
	out.Reset()
	if err := NewBrctl().Run(context.Background(), stdio, []string{"bogus"}); err == nil {
		t.Error("expected validation error for bad command")
	}
}

// TestHelpNotes asserts every capability-gated netctl applet documents a Notes
// section in --help (GitHub issues #712, #714, #716, #718, #719, #720).
func TestHelpNotes(t *testing.T) {
	t.Parallel()
	for _, c := range []*Command{
		NewBrctl(), NewIfenslave(), NewTunctl(),
		NewVconfig(), NewZcip(), NewNbdClient(),
	} {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		if err := c.Run(context.Background(), io, []string{"--help"}); err != nil {
			t.Fatalf("%s --help err = %v", c.Name(), err)
		}
		if !strings.Contains(out.String(), "Notes:") {
			t.Errorf("%s --help missing Notes section: %q", c.Name(), out.String())
		}
	}
}
