package halt_test

import (
	"bytes"
	"context"
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
