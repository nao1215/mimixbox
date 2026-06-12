package su

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

type runCall struct {
	acc  account
	argv []string
}

func setup(t *testing.T, root bool, authOK bool) *runCall {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "passwd")
	content := "root:x:0:0:root:/root:/bin/sh\nalice:x:1000:1000:alice:/home/alice:/bin/bash\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	call := &runCall{}
	op, oir, oaf, orf := passwdPath, isRootFn, authFn, runFn
	passwdPath = p
	isRootFn = func() bool { return root }
	authFn = func(string, string) (bool, error) { return authOK, nil }
	runFn = func(_ command.IO, acc account, argv []string) error {
		*call = runCall{acc, argv}
		return nil
	}
	t.Cleanup(func() { passwdPath, isRootFn, authFn, runFn = op, oir, oaf, orf })
	return call
}

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestRootNeedsNoAuth(t *testing.T) {
	call := setup(t, true, false) // auth would fail, but root skips it
	if err := run(t, "", "alice"); err != nil {
		t.Fatal(err)
	}
	if call.acc.name != "alice" || call.acc.uid != 1000 || call.acc.shell != "/bin/bash" {
		t.Errorf("target = %+v", call.acc)
	}
	if call.argv[0] != "bash" {
		t.Errorf("argv0 = %q, want bash", call.argv[0])
	}
}

func TestNonRootAuthSuccess(t *testing.T) {
	call := setup(t, false, true)
	if err := run(t, "password\n", "alice"); err != nil {
		t.Fatal(err)
	}
	if call.acc.name != "alice" {
		t.Errorf("should have run as alice")
	}
}

func TestNonRootAuthFailure(t *testing.T) {
	call := setup(t, false, false)
	if err := run(t, "wrong\n", "alice"); err == nil {
		t.Errorf("a failed auth should fail")
	}
	if call.acc.name != "" {
		t.Errorf("runFn must not be called on auth failure")
	}
}

func TestDefaultsToRoot(t *testing.T) {
	call := setup(t, true, true)
	if err := run(t, ""); err != nil {
		t.Fatal(err)
	}
	if call.acc.name != "root" {
		t.Errorf("default target = %q, want root", call.acc.name)
	}
}

func TestCommandAndLogin(t *testing.T) {
	call := setup(t, true, true)
	if err := run(t, "", "-l", "-c", "id", "alice"); err != nil {
		t.Fatal(err)
	}
	// Login shell -> argv0 prefixed with '-'; -c appends the command.
	if call.argv[0] != "-bash" || len(call.argv) != 3 || call.argv[1] != "-c" || call.argv[2] != "id" {
		t.Errorf("argv = %v", call.argv)
	}
}

func TestBareDashIsLogin(t *testing.T) {
	call := setup(t, true, true)
	if err := run(t, "", "-", "alice"); err != nil {
		t.Fatal(err)
	}
	if call.argv[0] != "-bash" {
		t.Errorf("bare '-' should give a login shell, argv0 = %q", call.argv[0])
	}
}

func TestShellOverride(t *testing.T) {
	call := setup(t, true, true)
	if err := run(t, "", "-s", "/bin/zsh", "alice"); err != nil {
		t.Fatal(err)
	}
	if call.acc.shell != "/bin/zsh" || call.argv[0] != "zsh" {
		t.Errorf("shell override not applied: %+v %v", call.acc, call.argv)
	}
}

func TestUnknownUser(t *testing.T) {
	setup(t, true, true)
	if err := run(t, "", "ghost"); err == nil {
		t.Errorf("an unknown user should fail")
	}
}
