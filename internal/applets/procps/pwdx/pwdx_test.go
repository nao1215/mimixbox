package pwdx

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withProc(t *testing.T, pid, cwd string) {
	t.Helper()
	dir := t.TempDir()
	pdir := filepath.Join(dir, pid)
	if err := os.MkdirAll(pdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cwd, filepath.Join(pdir, "cwd")); err != nil {
		t.Fatal(err)
	}
	orig := procDir
	procDir = dir
	t.Cleanup(func() { procDir = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestPrintCwd(t *testing.T) {
	withProc(t, "1234", "/home/alice/work")
	out, err := run(t, "1234")
	if err != nil {
		t.Fatal(err)
	}
	if out != "1234: /home/alice/work" {
		t.Errorf("pwdx = %q", out)
	}
}

func TestInvalidPid(t *testing.T) {
	withProc(t, "1", "/")
	if _, err := run(t, "notapid"); err == nil {
		t.Errorf("invalid PID should fail")
	}
}

func TestMissingProcess(t *testing.T) {
	withProc(t, "1", "/")
	if _, err := run(t, "9999"); err == nil {
		t.Errorf("a missing process should fail")
	}
}

func TestNoArgs(t *testing.T) {
	if _, err := run(t); err == nil {
		t.Errorf("no PID should fail")
	}
}
