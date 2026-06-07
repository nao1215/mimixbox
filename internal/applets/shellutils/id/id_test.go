package id_test

import (
	"bytes"
	"context"
	"os"
	"os/user"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/id"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := id.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunUserID(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "-u")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := strings.TrimSpace(out)
	want := strconv.Itoa(os.Getuid())
	if got != want {
		t.Errorf("uid = %q, want %q", got, want)
	}
}

func TestRunDefault(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "uid=") {
		t.Errorf("default output = %q, want to contain %q", out, "uid=")
	}
}

func TestRunUserName(t *testing.T) {
	t.Parallel()
	cur, err := user.Current()
	if err != nil {
		t.Skipf("cannot determine current user: %v", err)
	}
	out, errOut, runErr := run(t, "-u", "-n")
	if runErr != nil {
		t.Fatalf("Run error = %v (stderr=%q)", runErr, errOut)
	}
	got := strings.TrimSpace(out)
	if got != cur.Username {
		t.Errorf("user name = %q, want %q", got, cur.Username)
	}
}
