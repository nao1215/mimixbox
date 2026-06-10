package rdev

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func withMounts(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "mounts")
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := procMounts
	procMounts = f
	t.Cleanup(func() { procMounts = orig })
}

func run(t *testing.T) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, nil)
	return strings.TrimSpace(out.String()), err
}

func TestRootDevice(t *testing.T) {
	withMounts(t, "proc /proc proc rw 0 0\n/dev/sda1 / ext4 rw 0 0\ntmpfs /tmp tmpfs rw 0 0\n")
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if out != "/dev/sda1 /" {
		t.Errorf("rdev = %q, want %q", out, "/dev/sda1 /")
	}
}

func TestNoRoot(t *testing.T) {
	withMounts(t, "proc /proc proc rw 0 0\n")
	if _, err := run(t); err == nil {
		t.Errorf("missing root mount should fail")
	}
}

func TestMissingMounts(t *testing.T) {
	orig := procMounts
	procMounts = "/no/such/mounts"
	defer func() { procMounts = orig }()
	if _, err := run(t); err == nil {
		t.Errorf("missing mounts file should fail")
	}
}
