package xargs_test

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/findutils/xargs"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := xargs.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := xargs.New()
	if got := c.Name(); got != "xargs" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestEchoDefault(t *testing.T) {
	t.Parallel()
	// Default command is echo; all items end up on one line.
	out, errOut, err := run(t, "a b c\n", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestMaxArgs(t *testing.T) {
	t.Parallel()
	// -n 1 runs echo once per item, producing one line each.
	out, errOut, err := run(t, "x y z\n", "-n", "1", "echo")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got := strings.Fields(strings.TrimSpace(out))
	if len(got) != 3 {
		t.Errorf("out = %q, want three items on separate invocations", out)
	}
	if strings.Count(strings.TrimRight(out, "\n"), "\n") != 2 {
		t.Errorf("-n 1 should yield 3 lines, got %q", out)
	}
}

func TestReplace(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "world\n", "-I", "{}", "echo", "hello", "{}")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if strings.TrimSpace(out) != "hello world" {
		t.Errorf("out = %q, want 'hello world'", out)
	}
}

func TestNullDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a\x00b\x00c\x00", "-0", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestCustomDelimiter(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "a,b,c", "-d", ",", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "a b c" {
		t.Errorf("out = %q, want 'a b c'", out)
	}
}

func TestNoRunIfEmpty(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "   \n", "-r", "echo", "should-not-run")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty (command should not run)", out)
	}
}

func TestVerbose(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "hi\n", "-t", "echo")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(errOut, "echo hi") {
		t.Errorf("verbose stderr = %q, want 'echo hi'", errOut)
	}
}

func TestCommandFailure(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "this-command-does-not-exist-xyz")
	if err == nil {
		t.Error("expected error when command cannot be run")
	}
	if !strings.Contains(errOut, "xargs:") {
		t.Errorf("stderr = %q, want xargs: prefix", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: xargs") {
		t.Errorf("help = %q", out)
	}
}

func TestTrueCommandRunsOnEmptyWithoutR(t *testing.T) {
	t.Parallel()
	// Without -r, GNU xargs runs the command once even with empty input.
	// Use "true" (a real binary) to avoid output; on platforms without it,
	// skip.
	if runtime.GOOS != "linux" {
		t.Skip("relies on /usr/bin/true")
	}
	_, _, err := run(t, "", "true")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
}
