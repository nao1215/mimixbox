package fuser

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// setup builds a fixture /proc and a target file, with one process holding the
// target as an open fd and another using it as its cwd.
func setup(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	target := filepath.Join(base, "target")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	proc := filepath.Join(base, "proc")
	mklink := func(pid, rel, dest string) {
		p := filepath.Join(proc, pid, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(dest, p); err != nil {
			t.Fatal(err)
		}
	}
	mklink("100", "fd/3", target)  // process 100 has the file open
	mklink("200", "cwd", base)     // process 200's cwd is the directory
	mklink("300", "fd/4", "/etc/hosts") // process 300 uses something else

	orig := procDir
	procDir = proc
	t.Cleanup(func() { procDir = orig })
	return target
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestFindsOpenFile(t *testing.T) {
	target := setup(t)
	out, err := run(t, target)
	if err != nil {
		t.Fatal(err)
	}
	if out != "100" {
		t.Errorf("fuser = %q, want 100", out)
	}
}

func TestFindsByCwd(t *testing.T) {
	target := setup(t)
	dir := filepath.Dir(target)
	out, err := run(t, dir)
	if err != nil {
		t.Fatal(err)
	}
	if out != "200" {
		t.Errorf("fuser dir = %q, want 200", out)
	}
}

func TestNoUsers(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	unused := filepath.Join(dir, "lonely")
	if err := os.WriteFile(unused, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, unused); err == nil {
		t.Errorf("a file with no users should exit non-zero")
	}
}

func TestNoArgs(t *testing.T) {
	setup(t)
	if _, err := run(t); err == nil {
		t.Errorf("no file should fail")
	}
}
