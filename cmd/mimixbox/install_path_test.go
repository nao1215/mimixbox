package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

// stubSelf points resolveSelf at a fixed binary path so install decisions and
// symlink targets are deterministic regardless of the test host.
func stubSelf(t *testing.T, path string) {
	t.Helper()
	orig := osExecutable
	osExecutable = func() (string, error) { return path, nil }
	t.Cleanup(func() { osExecutable = orig })
}

func TestInstallEmptyDirDoesNotDependOnHostPath(t *testing.T) {
	// An empty target directory must receive applet symlinks regardless of
	// whether commands of the same name happen to exist on the host PATH
	// (issue #948, reproduction 1). "cat" exists on essentially every host, so
	// the old host-PATH gate would have skipped it.
	dir := t.TempDir()
	stubSelf(t, "/fresh/mimixbox")

	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install exit = %d (stderr: %s)", code, errBuf.String())
	}
	if fi, err := os.Lstat(filepath.Join(dir, "cat")); err != nil {
		t.Fatalf("expected cat symlink in an empty target dir: %v", err)
	} else if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("cat should be a symlink")
	}
}

func TestPlainInstallDoesNotOverwriteForeignSymlink(t *testing.T) {
	// --install must not replace a foreign symlink in the target directory
	// (issue #948, reproduction 2).
	dir := t.TempDir()
	foreign := filepath.Join(dir, "nyancat")
	if err := os.Symlink("/bin/true", foreign); err != nil {
		t.Fatal(err)
	}
	stubSelf(t, "/fresh/mimixbox")

	io, _, _ := newIO()
	if code := run([]string{"mimixbox", "--install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install exit = %d", code)
	}
	if got, _ := os.Readlink(foreign); got != "/bin/true" {
		t.Errorf("foreign symlink target = %q, want it left as /bin/true", got)
	}
}

func TestFullInstallDoesNotOverwriteForeignSymlink(t *testing.T) {
	// --full-install must not replace a foreign symlink either (issue #948,
	// reproduction 3).
	dir := t.TempDir()
	foreign := filepath.Join(dir, "cat")
	if err := os.Symlink("/bin/true", foreign); err != nil {
		t.Fatal(err)
	}
	stubSelf(t, "/fresh/mimixbox")

	io, _, _ := newIO()
	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d", code)
	}
	if got, _ := os.Readlink(foreign); got != "/bin/true" {
		t.Errorf("foreign symlink target = %q, want it left as /bin/true", got)
	}
}

func TestPlainInstallSkipsRealFileInTargetDir(t *testing.T) {
	// A real file occupying an applet name in the target directory must be left
	// untouched: ownership is decided by the target dir, not the host PATH.
	dir := t.TempDir()
	real := filepath.Join(dir, "cat")
	if err := os.WriteFile(real, []byte("not mimixbox"), 0o644); err != nil {
		t.Fatal(err)
	}
	stubSelf(t, "/fresh/mimixbox")

	io, _, _ := newIO()
	if code := run([]string{"mimixbox", "--install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install exit = %d", code)
	}
	if fi, err := os.Lstat(real); err != nil {
		t.Fatalf("real file vanished: %v", err)
	} else if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("real file was replaced by a symlink")
	}
	if data, _ := os.ReadFile(real); string(data) != "not mimixbox" {
		t.Errorf("real file content changed to %q", data)
	}
}
