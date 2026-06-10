package findfs

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
	label := filepath.Join(dir, "by-label")
	uuid := filepath.Join(dir, "by-uuid")
	for _, d := range []string{label, uuid} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("/dev/sda1", filepath.Join(label, "rootfs")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../sdb2", filepath.Join(uuid, "1234-5678")); err != nil {
		t.Fatal(err)
	}
	ol, ou := byLabelDir, byUUIDDir
	byLabelDir, byUUIDDir = label, uuid
	t.Cleanup(func() { byLabelDir, byUUIDDir = ol, ou })
}

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), err
}

func TestLabel(t *testing.T) {
	fixture(t)
	out, err := run(t, "LABEL=rootfs")
	if err != nil {
		t.Fatal(err)
	}
	if out != "/dev/sda1" {
		t.Errorf("findfs LABEL = %q, want /dev/sda1", out)
	}
}

func TestUUIDRelative(t *testing.T) {
	fixture(t)
	out, err := run(t, "UUID=1234-5678")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean(filepath.Join(byUUIDDir, "../../sdb2"))
	if out != want {
		t.Errorf("findfs UUID = %q, want %q", out, want)
	}
}

func TestCaseInsensitiveTag(t *testing.T) {
	fixture(t)
	if _, err := run(t, "label=rootfs"); err != nil {
		t.Errorf("lowercase tag should resolve: %v", err)
	}
}

func TestErrors(t *testing.T) {
	fixture(t)
	if _, err := run(t, "LABEL=missing"); err == nil {
		t.Errorf("missing label should fail")
	}
	if _, err := run(t, "BOGUS=x"); err == nil {
		t.Errorf("unknown tag kind should fail")
	}
	if _, err := run(t, "noequals"); err == nil {
		t.Errorf("malformed spec should fail")
	}
	if _, err := run(t); err == nil {
		t.Errorf("no argument should fail")
	}
}
