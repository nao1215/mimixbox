package readlink_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/readlink"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := readlink.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestReadSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	linkPath := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, linkPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != target+"\n" {
		t.Errorf("out = %q, want %q", out, target+"\n")
	}
}

func TestNoNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "t")
	linkPath := filepath.Join(dir, "l")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-n", linkPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != target {
		t.Errorf("out = %q, want %q", out, target)
	}
}

func TestCanonicalize(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	linkPath := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-f", linkPath)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	resolved, _ := filepath.EvalSymlinks(target)
	if strings.TrimRight(out, "\n") != resolved {
		t.Errorf("out = %q, want %q", out, resolved)
	}
}

func TestNotASymlink(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "plain")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := run(t, p)
	if err == nil {
		t.Fatal("expected error for a non-symlink")
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing operand") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := readlink.New()
	if c.Name() != "readlink" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
