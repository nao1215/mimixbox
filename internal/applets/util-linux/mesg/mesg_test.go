package mesg

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}

// withTTYFile makes resolveTTY return a temp file with the given mode, so the
// permission logic can be exercised without a real terminal.
func withTTYFile(t *testing.T, mode os.FileMode) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "tty")
	if err := os.WriteFile(f, nil, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(f, mode); err != nil {
		t.Fatal(err)
	}
	orig := resolveTTY
	resolveTTY = func(io.Reader) (string, error) { return f, nil }
	t.Cleanup(func() { resolveTTY = orig })
	return f
}

func run(t *testing.T, args ...string) (string, int) {
	t.Helper()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	err := New().Run(context.Background(), io, args)
	return strings.TrimSpace(out.String()), exitCode(err)
}

func TestReportAllowed(t *testing.T) {
	withTTYFile(t, 0o620) // group-writable
	out, code := run(t)
	if out != "is y" || code != 0 {
		t.Errorf("got %q code %d, want 'is y' 0", out, code)
	}
}

func TestReportDenied(t *testing.T) {
	withTTYFile(t, 0o600) // not group-writable
	out, code := run(t)
	if out != "is n" || code != 1 {
		t.Errorf("got %q code %d, want 'is n' 1", out, code)
	}
}

func TestEnable(t *testing.T) {
	f := withTTYFile(t, 0o600)
	if _, code := run(t, "y"); code != 0 {
		t.Errorf("mesg y code = %d, want 0", code)
	}
	info, _ := os.Stat(f)
	if info.Mode()&groupWrite == 0 {
		t.Errorf("group-write bit should be set, mode = %o", info.Mode())
	}
}

func TestDisable(t *testing.T) {
	f := withTTYFile(t, 0o620)
	if _, code := run(t, "n"); code != 1 {
		t.Errorf("mesg n code = %d, want 1", code)
	}
	info, _ := os.Stat(f)
	if info.Mode()&groupWrite != 0 {
		t.Errorf("group-write bit should be cleared, mode = %o", info.Mode())
	}
}

func TestInvalidArg(t *testing.T) {
	withTTYFile(t, 0o620)
	if _, code := run(t, "maybe"); code != 2 {
		t.Errorf("invalid arg code = %d, want 2", code)
	}
}

func TestNotATTY(t *testing.T) {
	// The real resolveTTY rejects a non-*os.File reader.
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := New().Run(context.Background(), io, nil)
	if exitCode(err) != 2 {
		t.Errorf("not-a-tty code = %d, want 2", exitCode(err))
	}
	if !strings.Contains(errBuf.String(), "terminal name") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}
