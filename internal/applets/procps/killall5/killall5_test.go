package killall5

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T, self int, cmdlines map[int]string) *map[int]syscall.Signal {
	t.Helper()
	dir := t.TempDir()
	for pid, cmd := range cmdlines {
		pdir := filepath.Join(dir, fmtInt(pid))
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pdir, "cmdline"), []byte(cmd), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	signaled := map[int]syscall.Signal{}
	origP, origS, origSig := procDir, selfPid, sendSignal
	procDir = dir
	selfPid = func() int { return self }
	sendSignal = func(pid int, sig syscall.Signal) error { signaled[pid] = sig; return nil }
	t.Cleanup(func() { procDir, selfPid, sendSignal = origP, origS, origSig })
	return &signaled
}

func fmtInt(n int) string { return strconv.Itoa(n) }

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: bytes.NewReader(nil), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func keys(m map[int]syscall.Signal) []int {
	var out []int
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func TestSignalsAllButExcluded(t *testing.T) {
	// PID 1 excluded, 2 is a kernel thread (empty cmdline), self is 100.
	sig := fixture(t, 100, map[int]string{1: "init", 2: "", 100: "test", 200: "bash", 300: "vim"})
	if err := run(t); err != nil {
		t.Fatal(err)
	}
	got := keys(*sig)
	if len(got) != 2 || got[0] != 200 || got[1] != 300 {
		t.Errorf("signaled = %v, want [200 300]", got)
	}
	if (*sig)[200] != syscall.SIGTERM {
		t.Errorf("default signal = %v, want SIGTERM", (*sig)[200])
	}
}

func TestSignalShorthand(t *testing.T) {
	sig := fixture(t, 100, map[int]string{200: "bash"})
	if err := run(t, "-9"); err != nil {
		t.Fatal(err)
	}
	if (*sig)[200] != syscall.SIGKILL {
		t.Errorf("killall5 -9 -> %v, want SIGKILL", (*sig)[200])
	}
}

func TestOmit(t *testing.T) {
	sig := fixture(t, 100, map[int]string{200: "bash", 300: "vim"})
	if err := run(t, "-o", "200"); err != nil {
		t.Fatal(err)
	}
	if _, ok := (*sig)[200]; ok {
		t.Errorf("PID 200 should be omitted")
	}
	if _, ok := (*sig)[300]; !ok {
		t.Errorf("PID 300 should be signaled")
	}
}

func TestNoTargets(t *testing.T) {
	fixture(t, 100, map[int]string{1: "init", 100: "self"})
	err := run(t)
	var ee *command.ExitError
	if e, ok := err.(*command.ExitError); ok {
		ee = e
	}
	if ee == nil || ee.Code != 2 {
		t.Errorf("err = %v, want exit 2", err)
	}
}
