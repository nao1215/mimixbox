package umount

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func setup(t *testing.T, unmountErr error) *string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "mounts")
	content := "/dev/sda1 / ext4 rw 0 0\n/dev/sdb1 /mnt/usb ext4 rw 0 0\n"
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	unmounted := new(string)
	om, ou := mountsPath, unmountFn
	mountsPath = f
	unmountFn = func(target string) error {
		*unmounted = target
		return unmountErr
	}
	t.Cleanup(func() { mountsPath, unmountFn = om, ou })
	return unmounted
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	io := command.IO{In: strings.NewReader(""), Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	return New().Run(context.Background(), io, args)
}

func TestUnmountByMountpoint(t *testing.T) {
	got := setup(t, nil)
	if err := run(t, "/mnt/usb"); err != nil {
		t.Fatal(err)
	}
	if *got != "/mnt/usb" {
		t.Errorf("unmounted %q, want /mnt/usb", *got)
	}
}

func TestUnmountByDevice(t *testing.T) {
	got := setup(t, nil)
	if err := run(t, "/dev/sdb1"); err != nil {
		t.Fatal(err)
	}
	if *got != "/mnt/usb" {
		t.Errorf("device unmount resolved to %q, want /mnt/usb", *got)
	}
}

func TestNotMounted(t *testing.T) {
	setup(t, nil)
	if err := run(t, "/not/mounted"); err == nil {
		t.Errorf("an unmounted target should fail")
	}
}

func TestUnmountFails(t *testing.T) {
	setup(t, errors.New("operation not permitted"))
	if err := run(t, "/mnt/usb"); err == nil {
		t.Errorf("an unmount failure should fail")
	}
}

func TestNoArg(t *testing.T) {
	setup(t, nil)
	if err := run(t); err == nil {
		t.Errorf("no target should fail")
	}
}
