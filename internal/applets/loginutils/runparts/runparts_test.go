package runparts

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func script(t *testing.T, dir, name, body string, exec bool) {
	t.Helper()
	mode := os.FileMode(0o644)
	if exec {
		mode = 0o755
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestRunsInOrder(t *testing.T) {
	dir := t.TempDir()
	script(t, dir, "20-second", "#!/bin/sh\necho second\n", true)
	script(t, dir, "10-first", "#!/bin/sh\necho first\n", true)
	out, err := run(t, dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != "first\nsecond\n" {
		t.Errorf("order = %q, want first then second", out)
	}
}

func TestSkipsNonExecAndInvalid(t *testing.T) {
	dir := t.TempDir()
	script(t, dir, "10-run", "#!/bin/sh\necho ran\n", true)
	script(t, dir, "20-noexec", "#!/bin/sh\necho nope\n", false)
	script(t, dir, "30-bad.bak", "#!/bin/sh\necho nope\n", true) // invalid name (dot)
	out, err := run(t, dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ran\n" {
		t.Errorf("output = %q, want only the valid executable", out)
	}
}

func TestListAndTest(t *testing.T) {
	dir := t.TempDir()
	script(t, dir, "10-exec", "#!/bin/sh\n", true)
	script(t, dir, "20-plain", "data\n", false)

	list, err := run(t, "--list", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(list, "10-exec") || !strings.Contains(list, "20-plain") {
		t.Errorf("--list should include both valid files:\n%s", list)
	}

	test, err := run(t, "--test", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(test, "10-exec") || strings.Contains(test, "20-plain") {
		t.Errorf("--test should include only the executable:\n%s", test)
	}
}

func TestPassesArg(t *testing.T) {
	dir := t.TempDir()
	script(t, dir, "10-echo", "#!/bin/sh\necho \"got:$1\"\n", true)
	out, err := run(t, "--arg", "hello", dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != "got:hello\n" {
		t.Errorf("arg not passed: %q", out)
	}
}

func TestFailurePropagates(t *testing.T) {
	dir := t.TempDir()
	script(t, dir, "10-fail", "#!/bin/sh\nexit 3\n", true)
	if _, err := run(t, dir); err == nil {
		t.Errorf("a failing program should make run-parts fail")
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("missing directory should fail")
	}
	if _, err := run(t, "/no/such/dir"); err == nil {
		t.Errorf("a missing directory should fail")
	}
}
