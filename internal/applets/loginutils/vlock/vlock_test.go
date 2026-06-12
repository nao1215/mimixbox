package vlock

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, authOK bool, authErr error) {
	t.Helper()
	ou, oa := currentUserFn, authFn
	currentUserFn = func() (string, error) { return "tester", nil }
	authFn = func(user, pw string) (bool, error) {
		if user != "tester" {
			t.Errorf("vlock must authenticate the current user, got %q", user)
		}
		return authOK, authErr
	}
	t.Cleanup(func() { currentUserFn, authFn = ou, oa })
}

func run(t *testing.T, stdin string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, nil)
	return out.String(), err
}

func TestUnlocksWithCorrectPassword(t *testing.T) {
	setup(t, true, nil)
	out, err := run(t, "secret\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "locked") || !strings.Contains(out, "unlocked") {
		t.Errorf("output = %q", out)
	}
}

func TestRejectsWrongPassword(t *testing.T) {
	setup(t, false, nil)
	if _, err := run(t, "wrong\n"); err == nil {
		t.Errorf("a wrong password should fail")
	}
}

func TestErrors(t *testing.T) {
	setup(t, false, errors.New("backend error"))
	if _, err := run(t, "x\n"); err == nil {
		t.Errorf("an auth backend error should fail")
	}
	setup(t, true, nil)
	if _, err := run(t, ""); err == nil {
		t.Errorf("no password should fail")
	}
}
