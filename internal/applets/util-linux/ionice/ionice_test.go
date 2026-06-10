package ionice

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type setCall struct {
	pid    int
	ioprio int
}

func withStubs(t *testing.T, have int) *setCall {
	t.Helper()
	captured := &setCall{pid: -999}
	origGet, origSet := ioprioGet, ioprioSet
	ioprioGet = func(int) (int, error) { return have, nil }
	ioprioSet = func(pid, ioprio int) error { captured.pid, captured.ioprio = pid, ioprio; return nil }
	t.Cleanup(func() { ioprioGet, ioprioSet = origGet, origSet })
	return captured
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func requireEcho(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skipf("echo not on PATH: %v", err)
	}
}

func TestPrintMode(t *testing.T) {
	withStubs(t, (classBestEffort<<ioprioClassShift)|4)
	out, err := run(t, "-p", "1234")
	if err != nil {
		t.Fatal(err)
	}
	if out != "best-effort: prio 4" {
		t.Errorf("out = %q", out)
	}
}

func TestSetPid(t *testing.T) {
	captured := withStubs(t, 0)
	if _, err := run(t, "-c", "3", "-p", "1234"); err != nil {
		t.Fatal(err)
	}
	if captured.pid != 1234 || captured.ioprio != classIdle<<ioprioClassShift {
		t.Errorf("set call = %+v", *captured)
	}
}

func TestRunCommand(t *testing.T) {
	requireEcho(t)
	captured := withStubs(t, 0)
	out, err := run(t, "-c", "2", "-n", "5", "echo", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if captured.pid != 0 || captured.ioprio != (classBestEffort<<ioprioClassShift)|5 {
		t.Errorf("set call = %+v", *captured)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("command output = %q", out)
	}
}

func TestSelfReport(t *testing.T) {
	withStubs(t, classIdle<<ioprioClassShift)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "idle" {
		t.Errorf("self report = %q", out)
	}
}

func TestValidation(t *testing.T) {
	withStubs(t, 0)
	if _, err := run(t, "-c", "2", "-n", "15", "echo"); err == nil {
		t.Errorf("out-of-range -n should fail")
	}
	if _, err := run(t, "-c", "9", "echo"); err == nil {
		t.Errorf("out-of-range -c should fail")
	}
	if _, err := run(t, "-c", "3", "-p", "1234", "--", "echo"); err == nil {
		t.Errorf("-p with a command should fail")
	}
}

func TestDescribe(t *testing.T) {
	t.Parallel()
	cases := map[int]string{
		(classRealtime << ioprioClassShift) | 2: "realtime: prio 2",
		classBestEffort << ioprioClassShift:     "best-effort: prio 0",
		classIdle << ioprioClassShift:           "idle",
		classNone << ioprioClassShift:           "none: prio 0",
	}
	for in, want := range cases {
		if got := describe(in); got != want {
			t.Errorf("describe(%#x) = %q, want %q", in, got, want)
		}
	}
}
