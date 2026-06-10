package flock

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestRunsCommandUnderLock(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "lock")
	out, err := run(t, lock, "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Errorf("output = %q, want hello", out)
	}
}

func TestRunsViaShell(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "lock")
	out, err := run(t, lock, "-c", "echo from-sh")
	if err != nil {
		t.Fatal(err)
	}
	if out != "from-sh" {
		t.Errorf("output = %q", out)
	}
}

func TestNonblockFailsWhenHeld(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "lock")
	// Hold an exclusive lock on the file from this test.
	f, err := os.OpenFile(lock, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		t.Fatal(err)
	}

	if _, err := run(t, "-n", lock, "echo", "should not run"); err == nil {
		t.Errorf("flock -n should fail when the lock is held")
	}
}

func TestExitCodePropagates(t *testing.T) {
	lock := filepath.Join(t.TempDir(), "lock")
	_, err := run(t, lock, "sh", "-c", "exit 3")
	ee, ok := err.(*command.ExitError)
	if !ok || ee.Code != 3 {
		t.Errorf("err = %v, want exit 3", err)
	}
}

func TestErrors(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("no FILE should fail")
	}
	lock := filepath.Join(t.TempDir(), "lock")
	if _, err := run(t, lock); err == nil {
		t.Errorf("no command should fail")
	}
}
