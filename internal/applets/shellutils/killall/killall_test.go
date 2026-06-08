package killall

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type killRecord struct {
	mu    sync.Mutex
	calls map[int]syscall.Signal
}

func (k *killRecord) record(pid int, sig syscall.Signal) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.calls == nil {
		k.calls = map[int]syscall.Signal{}
	}
	k.calls[pid] = sig
	return nil
}

func stub(t *testing.T, procs []process) *killRecord {
	t.Helper()
	origList, origKill := listProcesses, killProcess
	rec := &killRecord{}
	listProcesses = func() ([]process, error) { return procs, nil }
	killProcess = rec.record
	t.Cleanup(func() {
		listProcesses = origList
		killProcess = origKill
	})
	return rec
}

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func exitCode(err error) int {
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 0
}

var sample = []process{
	{pid: 300, name: "nginx"},
	{pid: 150, name: "nginx"},
	{pid: 100, name: "init"},
}

func TestKillsByName(t *testing.T) {
	rec := stub(t, sample)
	if _, _, err := run(t, "nginx"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if rec.calls[300] != syscall.SIGTERM || rec.calls[150] != syscall.SIGTERM {
		t.Errorf("calls = %v, want SIGTERM to 300 and 150", rec.calls)
	}
	if _, ok := rec.calls[100]; ok {
		t.Error("init should not have been signalled")
	}
}

func TestCustomSignal(t *testing.T) {
	rec := stub(t, sample)
	if _, _, err := run(t, "-s", "KILL", "nginx"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if rec.calls[300] != syscall.SIGKILL {
		t.Errorf("signal = %v, want SIGKILL", rec.calls[300])
	}
}

func TestNoMatch(t *testing.T) {
	stub(t, sample)
	_, errOut, err := run(t, "ghost")
	if code := exitCode(err); code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if !strings.Contains(errOut, "ghost: no process found") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestNoMatchQuiet(t *testing.T) {
	stub(t, sample)
	_, errOut, err := run(t, "-q", "ghost")
	if exitCode(err) != command.ExitFailure {
		t.Error("expected non-zero exit")
	}
	if errOut != "" {
		t.Errorf("quiet stderr = %q", errOut)
	}
}

func TestMissingArg(t *testing.T) {
	stub(t, sample)
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing process name") {
		t.Errorf("err = %v", err)
	}
}

func TestInvalidSignal(t *testing.T) {
	stub(t, sample)
	_, _, err := run(t, "-s", "BOGUS", "nginx")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown signal") {
		t.Errorf("err = %v", err)
	}
}

func TestRealProcfs(t *testing.T) {
	t.Parallel()
	procs, err := procFromProcfs()
	if err != nil {
		t.Fatalf("procFromProcfs() error = %v", err)
	}
	if len(procs) == 0 {
		t.Error("expected at least one process")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "killall" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
