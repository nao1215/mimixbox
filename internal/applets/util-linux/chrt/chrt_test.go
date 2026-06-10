package chrt

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type setCall struct {
	pid      int
	policy   int
	priority int
}

func withStubs(t *testing.T, policy, prio int) *setCall {
	t.Helper()
	captured := &setCall{pid: -999}
	origGS, origGP, origSS := getScheduler, getParam, setScheduler
	getScheduler = func(int) (int, error) { return policy, nil }
	getParam = func(int) (int, error) { return prio, nil }
	setScheduler = func(pid, p, pr int) error { *captured = setCall{pid, p, pr}; return nil }
	t.Cleanup(func() { getScheduler, getParam, setScheduler = origGS, origGP, origSS })
	return captured
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func requireEcho(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("echo"); err != nil {
		t.Skipf("echo not on PATH: %v", err)
	}
}

func TestPrintMode(t *testing.T) {
	withStubs(t, schedFIFO, 50)
	out, err := run(t, "-p", "1234")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "policy: SCHED_FIFO") || !strings.Contains(out, "priority: 50") {
		t.Errorf("print = %q", out)
	}
}

func TestSetPid(t *testing.T) {
	captured := withStubs(t, 0, 0)
	if _, err := run(t, "-r", "-p", "20", "1234"); err != nil {
		t.Fatal(err)
	}
	if *captured != (setCall{1234, schedRR, 20}) {
		t.Errorf("set call = %+v", *captured)
	}
}

func TestRunCommand(t *testing.T) {
	requireEcho(t)
	captured := withStubs(t, 0, 0)
	out, err := run(t, "-f", "30", "--", "echo", "go")
	if err != nil {
		t.Fatal(err)
	}
	if *captured != (setCall{0, schedFIFO, 30}) {
		t.Errorf("set call = %+v", *captured)
	}
	if !strings.Contains(out, "go") {
		t.Errorf("command output = %q", out)
	}
}

func TestErrors(t *testing.T) {
	withStubs(t, 0, 0)
	if _, err := run(t, "-o", "0"); err == nil {
		t.Errorf("missing command should fail")
	}
	if _, err := run(t, "-p", "notapid"); err == nil {
		t.Errorf("invalid PID should fail")
	}
	if _, err := run(t, "-f", "bad", "echo"); err == nil {
		t.Errorf("invalid priority should fail")
	}
}

func TestPolicyName(t *testing.T) {
	t.Parallel()
	cases := map[int]string{
		schedOther: "SCHED_OTHER", schedFIFO: "SCHED_FIFO", schedRR: "SCHED_RR",
		schedBatch: "SCHED_BATCH", schedIdle: "SCHED_IDLE",
	}
	for in, want := range cases {
		if got := policyName(in); got != want {
			t.Errorf("policyName(%d) = %q, want %q", in, got, want)
		}
	}
}
