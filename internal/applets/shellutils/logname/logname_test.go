package logname

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestPrintsLoginName(t *testing.T) {
	orig := loginName
	loginName = func() (string, error) { return "alice", nil }
	t.Cleanup(func() { loginName = orig })

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "alice\n" {
		t.Errorf("out = %q", out)
	}
}

func TestError(t *testing.T) {
	orig := loginName
	loginName = func() (string, error) { return "", errors.New("no name") }
	t.Cleanup(func() { loginName = orig })

	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no login name") {
		t.Errorf("err = %v", err)
	}
}

func TestCurrentLoginFromEnv(t *testing.T) {
	t.Setenv("LOGNAME", "bob")
	got, err := currentLogin()
	if err != nil {
		t.Fatalf("currentLogin() error = %v", err)
	}
	if got != "bob" {
		t.Errorf("got = %q, want bob", got)
	}
}

func TestCurrentLoginFallsBackToUser(t *testing.T) {
	t.Setenv("LOGNAME", "")
	t.Setenv("USER", "carol")
	got, err := currentLogin()
	if err != nil {
		t.Fatalf("currentLogin() error = %v", err)
	}
	if got != "carol" {
		t.Errorf("got = %q, want carol", got)
	}
}

func TestCurrentLoginFallsBackToCurrentUser(t *testing.T) {
	t.Setenv("LOGNAME", "")
	t.Setenv("USER", "")
	got, err := currentLogin()
	if err != nil {
		t.Fatalf("currentLogin() error = %v", err)
	}
	if got == "" {
		t.Error("expected a non-empty user name from user.Current()")
	}
}

func TestRealLognameRuns(t *testing.T) {
	t.Setenv("LOGNAME", "dave")
	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "dave\n" {
		t.Errorf("out = %q", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "logname" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
