package unlink_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/unlink"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := unlink.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRemoveFile(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Errorf("file still exists")
	}
}

func TestRejectDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, _, err := run(t, dir)
	if err == nil {
		t.Fatal("expected error unlinking a directory")
	}
}

func TestWrongArgCount(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exactly one argument") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "/no/such/file")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := unlink.New()
	if c.Name() != "unlink" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := unlink.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") || !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out.String())
	}
}
