package startstopdaemon

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type signalCall struct {
	pid int
	sig syscall.Signal
}

func stub(t *testing.T, running bool, startPID int, startErr error) (*[]string, *signalCall) {
	t.Helper()
	var started []string
	sig := &signalCall{pid: -1}
	oir, osp, osig := isRunning, startProc, signalProc
	isRunning = func(int) bool { return running }
	startProc = func(path string, args []string) (int, error) {
		started = append([]string{path}, args...)
		return startPID, startErr
	}
	signalProc = func(pid int, s syscall.Signal) error { *sig = signalCall{pid, s}; return nil }
	t.Cleanup(func() { isRunning, startProc, signalProc = oir, osp, osig })
	return &started, sig
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestStartWhenNotRunning(t *testing.T) {
	started, _ := stub(t, false, 4242, nil)
	pidfile := filepath.Join(t.TempDir(), "foo.pid")
	if err := run(t, "-S", "-p", pidfile, "-x", "/usr/bin/foo", "--", "arg1"); err != nil {
		t.Fatal(err)
	}
	if len(*started) != 2 || (*started)[0] != "/usr/bin/foo" || (*started)[1] != "arg1" {
		t.Errorf("started = %v", *started)
	}
	data, _ := os.ReadFile(pidfile)
	if strings.TrimSpace(string(data)) != "4242" {
		t.Errorf("pidfile = %q, want 4242", data)
	}
}

func TestStartWhenAlreadyRunning(t *testing.T) {
	started, _ := stub(t, true, 0, nil)
	pidfile := filepath.Join(t.TempDir(), "foo.pid")
	_ = os.WriteFile(pidfile, []byte("999\n"), 0o644)
	if err := run(t, "-S", "-p", pidfile, "-x", "/usr/bin/foo"); err != nil {
		t.Fatal(err)
	}
	if len(*started) != 0 {
		t.Errorf("should not start when already running, got %v", *started)
	}
}

func TestStop(t *testing.T) {
	_, sig := stub(t, true, 0, nil)
	pidfile := filepath.Join(t.TempDir(), "foo.pid")
	_ = os.WriteFile(pidfile, []byte("1234\n"), 0o644)
	if err := run(t, "-K", "-p", pidfile, "-s", "KILL"); err != nil {
		t.Fatal(err)
	}
	if sig.pid != 1234 || sig.sig != syscall.SIGKILL {
		t.Errorf("signal call = %+v, want pid 1234 SIGKILL", *sig)
	}
	if _, err := os.Stat(pidfile); !os.IsNotExist(err) {
		t.Errorf("pidfile should be removed after stop")
	}
}

func TestStopDefaultsToTerm(t *testing.T) {
	_, sig := stub(t, true, 0, nil)
	pidfile := filepath.Join(t.TempDir(), "foo.pid")
	_ = os.WriteFile(pidfile, []byte("7\n"), 0o644)
	if err := run(t, "-K", "-p", pidfile); err != nil {
		t.Fatal(err)
	}
	if sig.sig != syscall.SIGTERM {
		t.Errorf("default signal = %v, want SIGTERM", sig.sig)
	}
}

func TestStopNotRunning(t *testing.T) {
	stub(t, false, 0, nil)
	pidfile := filepath.Join(t.TempDir(), "foo.pid")
	_ = os.WriteFile(pidfile, []byte("1234\n"), 0o644)
	if err := run(t, "-K", "-p", pidfile); err == nil {
		t.Errorf("stopping a non-running process should fail")
	}
}

func TestModeAndArgErrors(t *testing.T) {
	stub(t, false, 1, nil)
	if err := run(t, "-p", "/tmp/x"); err == nil {
		t.Errorf("neither -S nor -K should fail")
	}
	if err := run(t, "-S", "-K"); err == nil {
		t.Errorf("both -S and -K should fail")
	}
	if err := run(t, "-S"); err == nil {
		t.Errorf("start without -x should fail")
	}
}

func TestParseSignal(t *testing.T) {
	t.Parallel()
	cases := map[string]syscall.Signal{"TERM": syscall.SIGTERM, "SIGKILL": syscall.SIGKILL, "9": syscall.Signal(9), "hup": syscall.SIGHUP}
	for in, want := range cases {
		if got, err := parseSignal(in); err != nil || got != want {
			t.Errorf("parseSignal(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if _, err := parseSignal("BOGUS"); err == nil {
		t.Errorf("unknown signal should error")
	}
}
