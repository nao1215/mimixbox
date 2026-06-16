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

func TestInstallReplacesExistingSymlink(t *testing.T) {
	// When a stale symlink already occupies an applet's name, full-install must
	// delete it and recreate it, reporting both actions.
	dir := t.TempDir()
	stale := filepath.Join(dir, "cat")
	if err := os.Symlink("/some/old/target", stale); err != nil {
		t.Fatal(err)
	}

	const wantTarget = "/fresh/mimixbox"
	orig := osExecutable
	osExecutable = func() (string, error) { return wantTarget, nil }
	t.Cleanup(func() { osExecutable = orig })

	io, out, errBuf := newIO()
	if code := run([]string{"mimixbox", "--full-install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("full-install exit = %d (stderr: %s)", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "Delete              : "+stale) {
		t.Errorf("stdout should report deleting the stale symlink, got %q", out.String())
	}
	target, err := os.Readlink(stale)
	if err != nil {
		t.Fatal(err)
	}
	if target != wantTarget {
		t.Errorf("recreated symlink target = %q, want %q", target, wantTarget)
	}
}

func TestPlainInstallSkipsExistingSystemCommand(t *testing.T) {
	// With -i/--install (full=false), an applet whose name collides with a
	// command already on the system is skipped with a warning, and no symlink is
	// created for it. "ls" is essentially always present on the test host.
	if _, err := os.Stat("/bin/ls"); err != nil {
		if _, err2 := os.Stat("/usr/bin/ls"); err2 != nil {
			t.Skip("ls not found on this host; cannot exercise the skip-existing branch")
		}
	}
	dir := t.TempDir()
	io, _, errBuf := newIO()
	if code := run([]string{"mimixbox", "--install", dir}, io); code != command.ExitSuccess {
		t.Fatalf("install exit = %d (stderr: %s)", code, errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "already exists. Not create symbolic link.") {
		t.Errorf("stderr should warn about at least one existing command, got %q", errBuf.String())
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
