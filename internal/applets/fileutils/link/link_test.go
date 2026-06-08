package link_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/link"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := link.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestCreateHardLink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, err := os.ReadFile(dst) //nolint:gosec // reading a file the test created
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "data" {
		t.Errorf("dst = %q", got)
	}
}

func TestWrongArgCount(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "only-one")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exactly two arguments") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingSource(t *testing.T) {
	t.Parallel()
	dst := filepath.Join(t.TempDir(), "dst")
	_, _, err := run(t, "/no/such/file", dst)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := link.New()
	if c.Name() != "link" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
