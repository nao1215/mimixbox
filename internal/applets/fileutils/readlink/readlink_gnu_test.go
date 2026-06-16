package readlink_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCanonicalizeExistingFailsOnMissing verifies -e fails when the final
// component does not exist.
func TestCanonicalizeExistingFailsOnMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope")
	_, _, err := run(t, "-e", missing)
	if err == nil {
		t.Fatal("expected -e to fail on a missing path")
	}
}

// TestCanonicalizeExistingSucceedsOnExisting verifies -e succeeds and prints
// the resolved path when every component exists.
func TestCanonicalizeExistingSucceedsOnExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "t")
	link := filepath.Join(dir, "l")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-e", link)
	if err != nil {
		t.Fatalf("-e on existing err = %v", err)
	}
	resolved, _ := filepath.EvalSymlinks(target)
	if strings.TrimRight(out, "\n") != resolved {
		t.Errorf("out = %q, want %q", out, resolved)
	}
}

// TestCanonicalizeMissingSucceeds verifies -m canonicalizes a path whose final
// components do not exist, without failing.
func TestCanonicalizeMissingSucceeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "a", "b", "c")
	out, _, err := run(t, "-m", missing)
	if err != nil {
		t.Fatalf("-m err = %v", err)
	}
	got := strings.TrimRight(out, "\n")
	// The existing prefix (dir) is resolved; the missing tail is appended.
	resolvedDir, _ := filepath.EvalSymlinks(dir)
	want := filepath.Join(resolvedDir, "a", "b", "c")
	if got != want {
		t.Errorf("out = %q, want %q", got, want)
	}
}

// TestZeroNulTerminator verifies -z terminates output with NUL, not newline.
func TestZeroNulTerminator(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "t")
	link := filepath.Join(dir, "l")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-z", link)
	if err != nil {
		t.Fatalf("-z err = %v", err)
	}
	if !strings.HasSuffix(out, "\x00") {
		t.Errorf("output %q should end with NUL", out)
	}
	if strings.HasSuffix(out, "\n") {
		t.Errorf("output %q should not end with newline under -z", out)
	}
	if out != target+"\x00" {
		t.Errorf("out = %q, want %q", out, target+"\x00")
	}
}
