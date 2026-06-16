package zip_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestZipVerbose covers addFile's verbose branch: each added entry is announced
// on stderr, and the archive still round-trips correctly.
func TestZipVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")

	_, errOut, err := run(t, "-v", archive, a, b)
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(errOut, "adding:") {
		t.Errorf("stderr = %q, want 'adding:' lines with -v", errOut)
	}
	got := names(t, archive)
	if got[filepath.ToSlash(a)] != "alpha" || got[filepath.ToSlash(b)] != "beta" {
		t.Errorf("round-trip mismatch: %v", got)
	}
}

// TestZipRecurseVerbose drives addDir together with the verbose path, archiving
// a nested tree.
func TestZipRecurseVerbose(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "tree", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "x.txt"), []byte("ex"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "out.zip")

	_, errOut, err := run(t, "-r", "-v", archive, filepath.Join(dir, "tree"))
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(errOut, "adding:") {
		t.Errorf("stderr = %q, want 'adding:' lines", errOut)
	}
	got := names(t, archive)
	if len(got) != 1 {
		t.Errorf("entries = %v, want 1 recursed file", got)
	}
}

// TestZipUnreadableFile covers addFile's os.Open error branch: a file that
// exists at stat time but cannot be opened for reading makes zip fail.
func TestZipUnreadableFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses file permission checks")
	}
	dir := t.TempDir()
	secret := filepath.Join(dir, "secret")
	if err := os.WriteFile(secret, []byte("nope"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(secret, 0o644) })
	archive := filepath.Join(dir, "out.zip")

	_, errOut, err := run(t, archive, secret)
	if err == nil {
		t.Fatal("expected error opening an unreadable file")
	}
	if !strings.Contains(errOut, "zip:") {
		t.Errorf("stderr = %q, want zip: prefix", errOut)
	}
}
