//go:build linux

package setfattr

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
	"golang.org/x/sys/unix"
)

// TestRealRoundTrip exercises the actual osBackend against a temp file. If the
// underlying filesystem does not support user xattrs (e.g. tmpfs without the
// option, or an overlay), it skips rather than failing.
func TestRealRoundTrip(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Probe support first.
	if err := unix.Setxattr(file, "user.probe", []byte("1"), 0); err != nil {
		if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EOPNOTSUPP) || errors.Is(err, unix.EPERM) {
			t.Skipf("filesystem does not support user xattrs: %v", err)
		}
		t.Fatalf("xattr probe failed: %v", err)
	}
	_ = unix.Removexattr(file, "user.probe")

	// Use the real backend (the package default) to set an attribute.
	var out, errBuf bytes.Buffer
	stdio := command.IO{In: strings.NewReader(""), Out: &out, Err: &errBuf}
	if err := New().Run(context.Background(), stdio, []string{"-n", "user.demo", "-v", "hello", file}); err != nil {
		t.Fatalf("setfattr failed: %v (stderr=%q)", err, errBuf.String())
	}

	// Verify with a direct syscall read.
	buf := make([]byte, 64)
	n, err := unix.Getxattr(file, "user.demo", buf)
	if err != nil {
		t.Fatalf("getxattr failed: %v", err)
	}
	if got := string(buf[:n]); got != "hello" {
		t.Errorf("round-trip value = %q, want %q", got, "hello")
	}

	// Remove it through the applet and confirm it is gone.
	errBuf.Reset()
	if err := New().Run(context.Background(), stdio, []string{"-x", "user.demo", file}); err != nil {
		t.Fatalf("setfattr -x failed: %v (stderr=%q)", err, errBuf.String())
	}
	if _, err := unix.Getxattr(file, "user.demo", buf); err == nil {
		t.Error("attribute still present after removal")
	}
}
