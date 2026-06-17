package taskset

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"github.com/nao1215/mimixbox/internal/testutil/fakecmd"
	"golang.org/x/sys/unix"
)

// requireEcho installs a repo-local fake echo on PATH so the exec path runs
// deterministically without depending on a host /bin/echo.
func requireEcho(t *testing.T) {
	t.Helper()
	fakecmd.UseOnly(t, "echo")
}

// withStubs makes getAffinity report `have` and captures whatever setAffinity is
// given.
func withStubs(t *testing.T, have *unix.CPUSet) **unix.CPUSet {
	t.Helper()
	var captured *unix.CPUSet
	origGet, origSet := getAffinity, setAffinity
	getAffinity = func(_ int, set *unix.CPUSet) error {
		if have != nil {
			*set = *have
		}
		return nil
	}
	setAffinity = func(_ int, set *unix.CPUSet) error { cp := *set; captured = &cp; return nil }
	t.Cleanup(func() { getAffinity, setAffinity = origGet, origSet })
	return &captured
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func cpuSet(cpus ...int) *unix.CPUSet {
	var s unix.CPUSet
	for _, c := range cpus {
		s.Set(c)
	}
	return &s
}

func TestPrintAffinity(t *testing.T) {
	withStubs(t, cpuSet(0, 1, 2, 3)) // mask 0xf
	out, err := run(t, "-p", "1234")
	if err != nil {
		t.Fatal(err)
	}
	if out != "pid 1234's current affinity mask: f\n" {
		t.Errorf("out = %q", out)
	}
}

func TestRunWithCPUList(t *testing.T) {
	requireEcho(t)
	captured := withStubs(t, nil)
	if _, err := run(t, "-c", "0,2-3", "echo", "ok"); err != nil {
		t.Fatal(err)
	}
	got := *captured
	if got == nil {
		t.Fatal("setAffinity not called")
	}
	for _, c := range []int{0, 2, 3} {
		if !got.IsSet(c) {
			t.Errorf("CPU %d should be set", c)
		}
	}
	if got.IsSet(1) {
		t.Errorf("CPU 1 should not be set")
	}
}

func TestRunWithHexMask(t *testing.T) {
	requireEcho(t)
	captured := withStubs(t, nil)
	if _, err := run(t, "0x5", "echo", "x"); err != nil { // bits 0 and 2
		t.Fatal(err)
	}
	got := *captured
	if !got.IsSet(0) || !got.IsSet(2) || got.IsSet(1) {
		t.Errorf("0x5 -> wrong CPUs")
	}
}

func TestSetPidAffinity(t *testing.T) {
	captured := withStubs(t, nil)
	out, err := run(t, "-p", "0x3", "999")
	if err != nil {
		t.Fatal(err)
	}
	got := *captured
	if !got.IsSet(0) || !got.IsSet(1) {
		t.Errorf("set affinity wrong")
	}
	if !strings.Contains(out, "new affinity mask: 3") {
		t.Errorf("out = %q", out)
	}
}

func TestErrors(t *testing.T) {
	withStubs(t, nil)
	if _, err := run(t, "zzz", "echo"); err == nil {
		t.Errorf("invalid mask should fail")
	}
	if _, err := run(t, "-p", "notapid"); err == nil {
		t.Errorf("invalid PID should fail")
	}
	if _, err := run(t, "0x1"); err == nil {
		t.Errorf("missing command should fail")
	}
}

func TestParseSet(t *testing.T) {
	t.Parallel()
	s, err := parseSet("0,2-4", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []int{0, 2, 3, 4} {
		if !s.IsSet(c) {
			t.Errorf("CPU %d missing", c)
		}
	}
	if _, err := parseSet("1-0", true); err == nil {
		t.Errorf("reversed range should fail")
	}
	if _, err := parseSet("xy", false); err == nil {
		t.Errorf("bad hex should fail")
	}
}
