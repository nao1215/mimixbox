package pidof

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func stub(t *testing.T, procs []process) {
	t.Helper()
	orig := listProcesses
	listProcesses = func() ([]process, error) { return procs, nil }
	t.Cleanup(func() { listProcesses = orig })
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
	{pid: 200, name: "bash"},
	{pid: 150, name: "nginx"},
	{pid: 100, name: "init"},
}

func TestMatchesByName(t *testing.T) {
	stub(t, sample)
	out, _, err := run(t, "nginx")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "300 150\n" {
		t.Errorf("out = %q, want %q", out, "300 150\n")
	}
}

func TestSingleShot(t *testing.T) {
	stub(t, sample)
	out, _, err := run(t, "-s", "nginx")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "300\n" {
		t.Errorf("out = %q", out)
	}
}

func TestMatchesByBasename(t *testing.T) {
	stub(t, sample)
	out, _, err := run(t, "/usr/sbin/nginx")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "300 150\n" {
		t.Errorf("out = %q", out)
	}
}

func TestNoMatchExitsOne(t *testing.T) {
	stub(t, sample)
	out, _, err := run(t, "no-such-prog")
	if code := exitCode(err); code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

func TestMissingArg(t *testing.T) {
	stub(t, sample)
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing program name") {
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
	if c.Name() != "pidof" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
