package mknod_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/mknod"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := mknod.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := mknod.New()
	if got := c.Name(); got != "mknod" {
		t.Errorf("Name() = %q, want %q", got, "mknod")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunFifo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fifo := filepath.Join(dir, "pipe")

	if _, errOut, err := run(t, "-m", "640", fifo, "p"); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	info, err := os.Stat(fifo)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeNamedPipe == 0 {
		t.Errorf("%s is not a FIFO (mode=%v)", fifo, info.Mode())
	}
	if info.Mode().Perm() != 0o640 {
		t.Errorf("mode = %o, want 640", info.Mode().Perm())
	}
}

func TestRunFifoWithDeviceNumbersFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "pipe"), "p", "1", "2")
	if err == nil {
		t.Error("expected error: FIFO must not have device numbers")
	}
	if !strings.Contains(errOut, "must not have device numbers") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunCharDeviceNeedsNumbers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "dev"), "c")
	if err == nil {
		t.Error("expected error: char device requires MAJOR and MINOR")
	}
	if !strings.Contains(errOut, "requires MAJOR and MINOR") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no operand", nil, "missing operand"},
		{"only name", []string{"foo"}, "missing operand"},
		{"invalid type", []string{"foo", "z"}, "invalid device type"},
		{"invalid mode", []string{"-m", "9z9", "foo", "p"}, "invalid mode"},
		{"invalid major", []string{"foo", "c", "xx", "2"}, "invalid major"},
		{"invalid minor", []string{"foo", "b", "1", "yy"}, "invalid minor"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errOut, err := run(t, tt.args...)
			if err == nil {
				t.Errorf("expected error for %v", tt.args)
			}
			if !strings.Contains(errOut, tt.want) {
				t.Errorf("stderr = %q, want substring %q", errOut, tt.want)
			}
		})
	}
}

// TestRunCharDeviceWithoutPrivilege verifies that creating a real device node
// without privileges fails cleanly. When the suite runs as root the operation
// may succeed, so the node is just cleaned up instead.
func TestRunCharDeviceWithoutPrivilege(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dev := filepath.Join(dir, "null")
	_, errOut, err := run(t, dev, "c", "1", "3")
	if os.Geteuid() == 0 {
		if err != nil {
			t.Fatalf("running as root, expected success, got %v (%s)", err, errOut)
		}
		_ = os.Remove(dev)
		return
	}
	if err == nil {
		t.Error("expected permission error creating device as non-root")
	}
}

func TestRunHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out, "Usage: mknod") {
		t.Errorf("help = %q, want usage line", out)
	}
}
