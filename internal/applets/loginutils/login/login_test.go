package login

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, authOK bool) *account {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "passwd")
	content := "root:x:0:0:root:/root:/bin/sh\nalice:x:1000:1000:alice:/home/alice:/bin/bash\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ran := &account{}
	op, oa, orf := passwdPath, authFn, runFn
	passwdPath = p
	authFn = func(string, string) (bool, error) { return authOK, nil }
	runFn = func(_ command.IO, acc account) error { *ran = acc; return nil }
	t.Cleanup(func() { passwdPath, authFn, runFn = op, oa, orf })
	return ran
}

func run(t *testing.T, stdin string, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(stdin), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestLoginWithOperand(t *testing.T) {
	ran := setup(t, true)
	if err := run(t, "secret\n", "alice"); err != nil {
		t.Fatal(err)
	}
	if ran.name != "alice" || ran.uid != 1000 || ran.shell != "/bin/bash" {
		t.Errorf("logged in account = %+v", *ran)
	}
}

func TestLoginUsernameFromStdin(t *testing.T) {
	ran := setup(t, true)
	if err := run(t, "alice\nsecret\n"); err != nil {
		t.Fatal(err)
	}
	if ran.name != "alice" {
		t.Errorf("username from stdin = %q", ran.name)
	}
}

func TestWrongPassword(t *testing.T) {
	ran := setup(t, false)
	if err := run(t, "wrong\n", "alice"); err == nil {
		t.Errorf("a wrong password should fail")
	}
	if ran.name != "" {
		t.Errorf("the shell must not start on a bad login")
	}
}

func TestForceSkipsAuth(t *testing.T) {
	ran := setup(t, false) // auth would fail, but -f skips it
	if err := run(t, "", "-f", "root"); err != nil {
		t.Fatal(err)
	}
	if ran.name != "root" {
		t.Errorf("-f should log in root without auth")
	}
}

func TestUnknownUser(t *testing.T) {
	setup(t, true)
	if err := run(t, "pw\n", "ghost"); err == nil {
		t.Errorf("an unknown user should fail")
	}
}
