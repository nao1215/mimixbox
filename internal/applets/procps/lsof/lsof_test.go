package lsof

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func fixture(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	pdir := filepath.Join(dir, "1234")
	if err := os.MkdirAll(filepath.Join(pdir, "fd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pdir, "comm"), []byte("myproc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mklink := func(rel, dest string) {
		if err := os.Symlink(dest, filepath.Join(pdir, rel)); err != nil {
			t.Fatal(err)
		}
	}
	mklink("cwd", "/work/dir")
	mklink("exe", "/usr/bin/myproc")
	mklink("fd/0", "/dev/null")
	mklink("fd/3", "/var/log/app.log")

	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return out.String(), err
}

func TestListProcess(t *testing.T) {
	fixture(t)
	out, err := run(t, "-p", "1234")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"COMMAND", "myproc", "1234",
		"cwd /work/dir", "txt /usr/bin/myproc",
		"0   /dev/null", "3   /var/log/app.log",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestListAll(t *testing.T) {
	fixture(t)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "myproc") {
		t.Errorf("all-processes listing missing myproc:\n%s", out)
	}
}

func TestMissingPid(t *testing.T) {
	fixture(t)
	if _, err := run(t, "-p", "9999"); err == nil {
		t.Errorf("a missing PID should fail")
	}
}
