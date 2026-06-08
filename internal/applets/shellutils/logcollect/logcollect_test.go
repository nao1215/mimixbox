package logcollect

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
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

func makeTree(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "syslog"), []byte("line1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "nginx"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "nginx", "access.log"), []byte("GET /\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return src
}

func TestCollectCopiesTree(t *testing.T) {
	t.Parallel()
	src := makeTree(t)
	dst := filepath.Join(t.TempDir(), "out")
	out, _, err := run(t, "-o", dst, src)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "collected 2 files") {
		t.Errorf("out = %q", out)
	}
	for _, rel := range []string{"syslog", "nginx/access.log"} {
		if _, err := os.Stat(filepath.Join(dst, rel)); err != nil {
			t.Errorf("expected %s to be copied: %v", rel, err)
		}
	}
}

func TestCollectPreservesContent(t *testing.T) {
	t.Parallel()
	src := makeTree(t)
	dst := filepath.Join(t.TempDir(), "out")
	if _, _, err := run(t, "-o", dst, src); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "syslog"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "line1\n" {
		t.Errorf("content = %q", got)
	}
}

func TestSourceMissing(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-o", t.TempDir(), "/no/such/source")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot read source") {
		t.Errorf("err = %v", err)
	}
}

func TestSourceNotDir(t *testing.T) {
	t.Parallel()
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := run(t, "-o", t.TempDir(), f)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("err = %v", err)
	}
}

func TestSkipsUnreadableFile(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "ok.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	secret := filepath.Join(src, "secret.log")
	if err := os.WriteFile(secret, []byte("y"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o644) })

	dst := filepath.Join(t.TempDir(), "out")
	out, _, err := run(t, "-o", dst, src)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// As non-root the 0000 file cannot be opened, so it is skipped; as root it
	// is readable and copied. Either way ok.log must be collected.
	if _, err := os.Stat(filepath.Join(dst, "ok.log")); err != nil {
		t.Errorf("ok.log should be collected: %v", err)
	}
	if !strings.Contains(out, "collected") {
		t.Errorf("out = %q", out)
	}
}

func TestDefaultOutputAndSource(t *testing.T) {
	t.Parallel()
	// Use an explicit empty source dir but the default output directory, then
	// clean it up. This exercises the default -o path.
	src := t.TempDir()
	wd, _ := os.Getwd()
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	out, _, err := run(t, src)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "collected-logs") {
		t.Errorf("expected default output dir in %q", out)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := New()
	if c.Name() != "log-collect" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
