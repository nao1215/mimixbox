package pwd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/pwd"
)

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := pwd.New()
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
	if c.Name() != "pwd" {
		t.Errorf("Name() = %q", c.Name())
	}
}

// TestLogicalIgnoresStalePWD covers the namesCurrentDir mismatch branch: when
// $PWD names a directory that is not the actual working directory, the logical
// path is discarded and os.Getwd() is used instead.
func TestLogicalIgnoresStalePWD(t *testing.T) {
	dir := t.TempDir()
	other := t.TempDir()
	t.Chdir(dir)
	// $PWD points somewhere else, so it must not be trusted.
	t.Setenv("PWD", other)

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	if sameDir(t, got, other) {
		t.Errorf("output %q should not be the stale PWD %q", got, other)
	}
	if !sameDir(t, got, dir) {
		t.Errorf("output = %q, want the real working directory %q", got, dir)
	}
}

// TestLogicalIgnoresRelativePWD covers the !filepath.IsAbs(pwd) branch: a
// relative $PWD is ignored entirely.
func TestLogicalIgnoresRelativePWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PWD", "relative/path")

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	if !filepath.IsAbs(got) {
		t.Errorf("output %q should be absolute despite a relative PWD", got)
	}
	if !sameDir(t, got, dir) {
		t.Errorf("output = %q, want %q", got, dir)
	}
}

// TestLogicalIgnoresNonexistentPWD covers the namesCurrentDir branch where
// EvalSymlinks of $PWD fails because the path does not exist.
func TestLogicalIgnoresNonexistentPWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PWD", filepath.Join(dir, "does", "not", "exist"))

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	if !sameDir(t, got, dir) {
		t.Errorf("output = %q, want %q", got, dir)
	}
}

// TestPhysicalThroughSymlink covers workingDir(physical=true) resolving a
// symlinked working directory to its canonical target.
func TestPhysicalThroughSymlink(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "real")
	link := filepath.Join(base, "link")
	if err := os.Mkdir(real, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
	t.Chdir(link)

	out, _, err := run(t, "-P")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	resolved, err := filepath.EvalSymlinks(real)
	if err != nil {
		t.Fatal(err)
	}
	if got != resolved {
		t.Errorf("-P output = %q, want resolved %q", got, resolved)
	}
}
