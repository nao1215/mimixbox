package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func TestRemoveRequiresDirOperand(t *testing.T) {
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--remove"}, io); code != command.ExitFailure {
		t.Fatalf("remove with no dir exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "single DIRECTORY operand") {
		t.Errorf("stderr = %q, want operand error", errBuf.String())
	}
}

func TestInstallNonexistentDirFails(t *testing.T) {
	// runInstall must surface install()'s "no such directory" error on stderr
	// and return a failure code.
	io, _, errBuf := newIO()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if code := run([]string{"mimixbox", "--full-install", missing}, io); code != command.ExitFailure {
		t.Fatalf("install into missing dir exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "no such directory") {
		t.Errorf("stderr = %q, want 'no such directory'", errBuf.String())
	}
}

func TestRemoveNonexistentDirFails(t *testing.T) {
	io, _, errBuf := newIO()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if code := run([]string{"mimixbox", "--remove", missing}, io); code != command.ExitFailure {
		t.Fatalf("remove from missing dir exit = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "no such directory") {
		t.Errorf("stderr = %q, want 'no such directory'", errBuf.String())
	}
}

func TestResolveSelfPrefersExecutable(t *testing.T) {
	const want = "/exact/running/mimixbox"
	orig := osExecutable
	osExecutable = func() (string, error) { return want, nil }
	t.Cleanup(func() { osExecutable = orig })

	// The argv[0] fallback ("anything") must be ignored when os.Executable works.
	got, err := resolveSelf("anything")
	if err != nil {
		t.Fatalf("resolveSelf error = %v", err)
	}
	if got != want {
		t.Errorf("resolveSelf = %q, want %q", got, want)
	}
}

func TestResolveSelfFallsBackToArgv0(t *testing.T) {
	// When os.Executable fails, resolveSelf falls back to the absolute path of
	// the invoked argv[0].
	orig := osExecutable
	osExecutable = func() (string, error) { return "", errors.New("no executable path") }
	t.Cleanup(func() { osExecutable = orig })

	got, err := resolveSelf("relative/mimixbox")
	if err != nil {
		t.Fatalf("resolveSelf error = %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("resolveSelf fallback = %q, want an absolute path", got)
	}
	if !strings.HasSuffix(got, filepath.Join("relative", "mimixbox")) {
		t.Errorf("resolveSelf fallback = %q, want it to end with the invoked path", got)
	}
}

func TestFullInstallRefreshesOwnSymlink(t *testing.T) {
	// When a symlink already owned by this MimixBox occupies an applet's name,
	// --full-install refreshes it: it deletes and recreates the link, reporting
	// both actions. (A foreign symlink, by contrast, must be left alone.)
	dir := t.TempDir()
	const wantTarget = "/fresh/mimixbox"
	own := filepath.Join(dir, "cat")
	if err := os.Symlink(wantTarget, own); err != nil {
		t.Fatal(err)
	}

	orig := osExecutable
	osExecutable = func() (string, error) { return wantTarget, nil }
	t.Cleanup(func() { osExecutable = orig })

	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d (stderr: %s)", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Delete              : "+own) {
		t.Errorf("stdout should report refreshing our own symlink, got %q", out.String())
	}
	target, err := os.Readlink(own)
	if err != nil {
		t.Fatal(err)
	}
	if target != wantTarget {
		t.Errorf("recreated symlink target = %q, want %q", target, wantTarget)
	}
}

func TestPlainInstallSkipsForeignEntry(t *testing.T) {
	// With -i/--install, an applet name already occupied by something not owned
	// by MimixBox is skipped with a warning, and the foreign entry is left in
	// place. The decision is based on the target directory, not the host PATH.
	dir := t.TempDir()
	foreign := filepath.Join(dir, "cat")
	if err := os.Symlink("/bin/true", foreign); err != nil {
		t.Fatal(err)
	}
	orig := osExecutable
	osExecutable = func() (string, error) { return "/fresh/mimixbox", nil }
	t.Cleanup(func() { osExecutable = orig })

	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install exit = %d (stderr: %s)", code, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "not owned by MimixBox") {
		t.Errorf("stderr should warn about the foreign entry, got %q", errBuf.String())
	}
	if got, _ := os.Readlink(foreign); got != "/bin/true" {
		t.Errorf("foreign symlink target = %q, want it left as /bin/true", got)
	}
}

func TestOwnedBySelfComparesCleanedPaths(t *testing.T) {
	if !ownedBySelf("/a/b/../b/mimixbox", "/a/b/mimixbox") {
		t.Errorf("equivalent path spellings should be owned by self")
	}
	if ownedBySelf("/opt/other/mimixbox", "/usr/bin/mimixbox") {
		t.Errorf("different targets must not be owned by self")
	}
}
