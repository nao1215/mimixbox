package renice

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type call struct {
	pid  int
	nice int
}

func withStubs(t *testing.T, old int) *[]call {
	t.Helper()
	var calls []call
	origGet, origSet := getNice, setNice
	getNice = func(int) (int, error) { return old, nil }
	setNice = func(pid, nice int) error { calls = append(calls, call{pid, nice}); return nil }
	t.Cleanup(func() { getNice, setNice = origGet, origSet })
	return &calls
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestPositionalPriority(t *testing.T) {
	calls := withStubs(t, 0)
	out, err := run(t, "5", "-p", "1234")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0] != (call{1234, 5}) {
		t.Errorf("calls = %v, want [{1234 5}]", *calls)
	}
	if !strings.Contains(out, "1234 (process ID) old priority 0, new priority 5") {
		t.Errorf("output = %q", out)
	}
}

func TestDashNPriority(t *testing.T) {
	calls := withStubs(t, 2)
	if _, err := run(t, "-n", "10", "1234", "5678"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 || (*calls)[0] != (call{1234, 10}) || (*calls)[1] != (call{5678, 10}) {
		t.Errorf("calls = %v", *calls)
	}
}

func TestInvalidPid(t *testing.T) {
	withStubs(t, 0)
	if _, err := run(t, "5", "notapid"); err == nil {
		t.Errorf("invalid PID should fail")
	}
}

func TestMissingPriority(t *testing.T) {
	withStubs(t, 0)
	if _, err := run(t); err == nil {
		t.Errorf("missing priority should fail")
	}
}

func TestMissingPid(t *testing.T) {
	withStubs(t, 0)
	if _, err := run(t, "5"); err == nil {
		t.Errorf("missing PID should fail")
	}
}
