package halt_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/pmutils/halt"
	"github.com/nao1215/mimixbox/internal/command"
)

// recorder captures the action passed to the stubbed reboot function instead of
// actually stopping the machine.
type recorder struct {
	called bool
	action int
}

func (r *recorder) reboot(action int) error {
	r.called = true
	r.action = action
	return nil
}

// run executes cmd with a fake-root environment and the reboot syscall stubbed
// out, returning the recorded action and the command result.
func run(t *testing.T, cmd *halt.Command, args ...string) (rec *recorder, out, errOut string, err error) {
	t.Helper()
	rec = &recorder{}
	halt.SetRebootFnForTest(t, rec.reboot)
	halt.SetIsRootForTest(t, func() bool { return true })
	halt.SetWtmpFileForTest(t, filepath.Join(t.TempDir(), "wtmp"))

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: outBuf, Err: errBuf}
	err = cmd.Run(context.Background(), io, args)
	return rec, outBuf.String(), errBuf.String(), err
}

func TestNameAndSynopsis(t *testing.T) {
	tests := []struct {
		cmd      *halt.Command
		name     string
		synopsis string
	}{
		{halt.NewHalt(), "halt", "Halt the system"},
		{halt.NewPoweroff(), "poweroff", "Power off the system"},
		{halt.NewReboot(), "reboot", "Reboot the system"},
	}
	for _, tt := range tests {
		if got := tt.cmd.Name(); got != tt.name {
			t.Errorf("Name() = %q, want %q", got, tt.name)
		}
		if got := tt.cmd.Synopsis(); got != tt.synopsis {
			t.Errorf("Synopsis() = %q, want %q", got, tt.synopsis)
		}
	}
}

func TestRunRequestsCorrectAction(t *testing.T) {
	tests := []struct {
		name   string
		cmd    *halt.Command
		action int
	}{
		{"halt", halt.NewHalt(), syscall.LINUX_REBOOT_CMD_HALT},
		{"poweroff", halt.NewPoweroff(), syscall.LINUX_REBOOT_CMD_POWER_OFF},
		{"reboot", halt.NewReboot(), syscall.LINUX_REBOOT_CMD_RESTART},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rec, _, _, err := run(t, tt.cmd)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if !rec.called {
				t.Fatal("reboot function was not called")
			}
			if rec.action != tt.action {
				t.Errorf("action = %#x, want %#x", rec.action, tt.action)
			}
		})
	}
}

func TestRunNotRoot(t *testing.T) {
	rec := &recorder{}
	halt.SetRebootFnForTest(t, rec.reboot)
	halt.SetIsRootForTest(t, func() bool { return false })

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: outBuf, Err: errBuf}
	err := halt.NewHalt().Run(context.Background(), io, nil)

	if err == nil {
		t.Fatal("expected error when not root")
	}
	if rec.called {
		t.Fatal("reboot function must NOT be called when not root")
	}
	if !strings.Contains(errBuf.String(), "root") {
		t.Errorf("stderr = %q, want a permission message mentioning root", errBuf.String())
	}
}

func TestRunHaltPoweroffOption(t *testing.T) {
	rec, _, _, err := run(t, halt.NewHalt(), "-p")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !rec.called {
		t.Fatal("reboot function was not called")
	}
	if rec.action != syscall.LINUX_REBOOT_CMD_POWER_OFF {
		t.Errorf("halt -p action = %#x, want POWER_OFF %#x", rec.action, syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}
}

func TestRunWtmpOnly(t *testing.T) {
	rec := &recorder{}
	halt.SetRebootFnForTest(t, rec.reboot)
	halt.SetIsRootForTest(t, func() bool { return true })
	wtmp := filepath.Join(t.TempDir(), "wtmp")
	halt.SetWtmpFileForTest(t, wtmp)

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := halt.NewReboot().Run(context.Background(), io, []string{"-w"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if rec.called {
		t.Fatal("-w (wtmp-only) must NOT stop the system")
	}
	info, statErr := os.Stat(wtmp)
	if statErr != nil {
		t.Fatalf("wtmp not written: %v", statErr)
	}
	if info.Size() != 384 {
		t.Errorf("wtmp size = %d, want 384 (one utmp record)", info.Size())
	}
}

func TestRunNoWtmp(t *testing.T) {
	rec := &recorder{}
	halt.SetRebootFnForTest(t, rec.reboot)
	halt.SetIsRootForTest(t, func() bool { return true })
	wtmp := filepath.Join(t.TempDir(), "wtmp")
	halt.SetWtmpFileForTest(t, wtmp)

	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if err := halt.NewHalt().Run(context.Background(), io, []string{"-d"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, statErr := os.Stat(wtmp); !os.IsNotExist(statErr) {
		t.Errorf("-d must not write wtmp, stat error = %v", statErr)
	}
}

func TestRunNoSync(t *testing.T) {
	synced := false
	halt.SetSyncFnForTest(t, func() { synced = true })

	if _, _, _, err := run(t, halt.NewHalt(), "-n"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if synced {
		t.Error("-n must not sync before halt")
	}

	synced = false
	if _, _, _, err := run(t, halt.NewHalt()); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !synced {
		t.Error("default halt must sync before halt")
	}
}

func TestRunHelp(t *testing.T) {
	rec := &recorder{}
	halt.SetRebootFnForTest(t, rec.reboot)
	halt.SetIsRootForTest(t, func() bool { return true })

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: outBuf, Err: errBuf}
	if err := halt.NewHalt().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if rec.called {
		t.Fatal("reboot function must NOT be called for --help")
	}
	if !strings.Contains(outBuf.String(), "Usage:") {
		t.Errorf("stdout = %q, want usage", outBuf.String())
	}
}

// TestHelpSections asserts reboot/poweroff/halt --help renders structured help.
func TestHelpSections(t *testing.T) {
	t.Parallel()
	for _, c := range []*halt.Command{halt.NewReboot(), halt.NewPoweroff(), halt.NewHalt()} {
		out := &bytes.Buffer{}
		io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
		if err := c.Run(context.Background(), io, []string{"--help"}); err != nil {
			t.Fatalf("%s --help err = %v", c.Name(), err)
		}
		for _, want := range []string{"Usage: " + c.Name(), "Examples:", "Exit status:"} {
			if !strings.Contains(out.String(), want) {
				t.Errorf("%s --help missing %q: %q", c.Name(), want, out.String())
			}
		}
	}
}
