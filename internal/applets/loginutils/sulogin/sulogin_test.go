package sulogin

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, authOK bool, authErr error) *string {
	t.Helper()
	ranShell := new(string)
	*ranShell = "<none>"
	oa, or := authFn, runFn
	authFn = func(user, pw string) (bool, error) {
		if user != "root" {
			t.Errorf("sulogin must authenticate root, got %q", user)
		}
		return authOK, authErr
	}
	runFn = func(_ command.IO, shell string) error { *ranShell = shell; return nil }
	t.Cleanup(func() { authFn, runFn = oa, or })
	return ranShell
}

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestStartsShellOnSuccess(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")
	ran := setup(t, true, nil)
	if err := run(t, "rootpw\n"); err != nil {
		t.Fatal(err)
	}
	if *ran != "/bin/bash" {
		t.Errorf("ran shell = %q, want /bin/bash", *ran)
	}
}

func TestRejectsWrongPassword(t *testing.T) {
	ran := setup(t, false, nil)
	if err := run(t, "wrong\n"); err == nil {
		t.Errorf("a wrong password should fail")
	}
	if *ran != "<none>" {
		t.Errorf("the shell must not start on auth failure")
	}
}

func TestDefaultShell(t *testing.T) {
	t.Setenv("SHELL", "")
	ran := setup(t, true, nil)
	if err := run(t, "pw\n"); err != nil {
		t.Fatal(err)
	}
	if *ran != "/bin/sh" {
		t.Errorf("default shell = %q, want /bin/sh", *ran)
	}
}

func TestErrors(t *testing.T) {
	setup(t, false, errors.New("backend error"))
	if err := run(t, "pw\n"); err == nil {
		t.Errorf("an auth backend error should fail")
	}
	setup(t, true, nil)
	if err := run(t, ""); err == nil { // no password line
		t.Errorf("missing password should fail")
	}
}
