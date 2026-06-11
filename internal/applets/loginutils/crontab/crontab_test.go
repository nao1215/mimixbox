package crontab

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "crontabs")
	od, ou := spoolDir, currentUserFn
	spoolDir = dir
	currentUserFn = func() (string, error) { return "tester", nil }
	t.Cleanup(func() { spoolDir, currentUserFn = od, ou })
	return dir
}

func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestInstallFromStdinThenList(t *testing.T) {
	dir := setup(t)
	const cron = "*/5 * * * * /usr/bin/backup\n"
	if _, err := run(t, cron); err != nil {
		t.Fatal(err)
	}
	// The spool file for "tester" must hold the content.
	data, err := os.ReadFile(filepath.Join(dir, "tester"))
	if err != nil || string(data) != cron {
		t.Fatalf("installed crontab = %q, err %v", data, err)
	}
	// -l prints it back.
	out, err := run(t, "", "-l")
	if err != nil {
		t.Fatal(err)
	}
	if out != cron {
		t.Errorf("-l output = %q", out)
	}
}

func TestInstallFromFile(t *testing.T) {
	setup(t)
	src := filepath.Join(t.TempDir(), "my.cron")
	_ = os.WriteFile(src, []byte("0 0 * * * /bin/true\n"), 0o644)
	if _, err := run(t, "", src); err != nil {
		t.Fatal(err)
	}
	out, _ := run(t, "", "-l")
	if !strings.Contains(out, "/bin/true") {
		t.Errorf("file not installed: %q", out)
	}
}

func TestRemove(t *testing.T) {
	dir := setup(t)
	if _, err := run(t, "1 2 3 4 5 cmd\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "", "-r"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tester")); !os.IsNotExist(err) {
		t.Errorf("crontab should be removed")
	}
}

func TestSpecificUser(t *testing.T) {
	dir := setup(t)
	if _, err := run(t, "@daily x\n", "-u", "alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "alice")); err != nil {
		t.Errorf("alice crontab not installed: %v", err)
	}
}

func TestErrors(t *testing.T) {
	setup(t)
	if _, err := run(t, "", "-l"); err == nil {
		t.Errorf("listing a missing crontab should fail")
	}
	if _, err := run(t, "", "-r"); err == nil {
		t.Errorf("removing a missing crontab should fail")
	}
	if _, err := run(t, "", "-e"); err == nil {
		t.Errorf("interactive edit should be unsupported")
	}
}
