package getty

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, issue string, loginErr error) *string {
	t.Helper()
	got := new(string)
	*got = "<none>"
	oi, ol := issuePath, loginExecFn
	if issue != "" {
		p := filepath.Join(t.TempDir(), "issue")
		if err := os.WriteFile(p, []byte(issue), 0o644); err != nil {
			t.Fatal(err)
		}
		issuePath = p
	} else {
		issuePath = filepath.Join(t.TempDir(), "no-issue")
	}
	loginExecFn = func(_ command.IO, username string) error { *got = username; return loginErr }
	t.Cleanup(func() { issuePath, loginExecFn = oi, ol })
	return got
}

func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestPromptsAndHandsOff(t *testing.T) {
	got := setup(t, "Welcome to mimixbox\n", nil)
	out, err := run(t, "alice\n", "38400", "tty1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Welcome to mimixbox") || !strings.Contains(out, "login: ") {
		t.Errorf("banner/prompt missing:\n%s", out)
	}
	if *got != "alice" {
		t.Errorf("handed off username %q, want alice", *got)
	}
}

func TestNoTTY(t *testing.T) {
	setup(t, "", nil)
	if _, err := run(t, "alice\n"); err == nil {
		t.Errorf("a missing TTY should fail")
	}
}

func TestEmptyUsername(t *testing.T) {
	got := setup(t, "", nil)
	if _, err := run(t, "\n", "tty1"); err == nil {
		t.Errorf("an empty username should fail")
	}
	if *got != "<none>" {
		t.Errorf("login must not be invoked without a username")
	}
}

func TestLoginFailure(t *testing.T) {
	setup(t, "", errors.New("login refused"))
	if _, err := run(t, "alice\n", "tty1"); err == nil {
		t.Errorf("a login failure should fail")
	}
}
