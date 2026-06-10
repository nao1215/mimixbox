package losetup

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
	// loop0 is active (has a backing file).
	loop0 := filepath.Join(dir, "loop0", "loop")
	if err := os.MkdirAll(loop0, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loop0, "backing_file"), []byte("/images/disk.img\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// loop1 is inactive (no loop/ subdirectory).
	if err := os.MkdirAll(filepath.Join(dir, "loop1"), 0o755); err != nil {
		t.Fatal(err)
	}
	// sda is not a loop device.
	if err := os.MkdirAll(filepath.Join(dir, "sda"), 0o755); err != nil {
		t.Fatal(err)
	}
	orig := sysBlockDir
	sysBlockDir = dir
	t.Cleanup(func() { sysBlockDir = orig })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestListsActive(t *testing.T) {
	fixture(t)
	out, err := run(t, "-a")
	if err != nil {
		t.Fatal(err)
	}
	if out != "/dev/loop0: (/images/disk.img)" {
		t.Errorf("losetup -a = %q", out)
	}
}

func TestNoArgListsAll(t *testing.T) {
	fixture(t)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "loop0") {
		t.Errorf("no-arg listing = %q", out)
	}
	if strings.Contains(out, "loop1") || strings.Contains(out, "sda") {
		t.Errorf("inactive/non-loop devices should be skipped: %q", out)
	}
}

func TestSetupUnsupported(t *testing.T) {
	fixture(t)
	if _, err := run(t, "/dev/loop0", "/tmp/img"); err == nil {
		t.Errorf("requesting a setup should fail deterministically")
	}
}
