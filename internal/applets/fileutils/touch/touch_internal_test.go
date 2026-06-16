package touch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestTouchAccessOnlyPreservesModTime exercises the -a branch: only the access
// time is advanced, the modification time keeps its old value.
func TestTouchAccessOnlyPreservesModTime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-3 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	if err := touch(path, options{accessOnly: true}); err != nil {
		t.Fatalf("touch -a err = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// -a must leave the modification time unchanged (within a second).
	if delta := info.ModTime().Sub(old); delta < -time.Second || delta > time.Second {
		t.Errorf("modtime moved by %v, want unchanged", delta)
	}
}

// TestTouchModifyOnlyAdvancesModTime exercises the -m branch.
func TestTouchModifyOnlyAdvancesModTime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-3 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	if err := touch(path, options{modifyOnly: true}); err != nil {
		t.Fatalf("touch -m err = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().After(old) {
		t.Errorf("modtime = %v, want advanced past %v", info.ModTime(), old)
	}
}

// TestTouchNoCreateOnExistingStillUpdates verifies -c does not block updating an
// already-existing file's timestamps.
func TestTouchNoCreateOnExistingStillUpdates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-3 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	if err := touch(path, options{noCreate: true}); err != nil {
		t.Fatalf("touch -c err = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().After(old) {
		t.Errorf("modtime = %v, want advanced", info.ModTime())
	}
}

// TestTouchNoCreateOnMissingIsNoop verifies -c leaves a missing file absent and
// returns no error.
func TestTouchNoCreateOnMissingIsNoop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.txt")
	if err := touch(path, options{noCreate: true}); err != nil {
		t.Fatalf("touch -c missing err = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should remain absent, stat err = %v", err)
	}
}

// TestTouchCreateFailsInMissingDir exercises the os.Create error path.
func TestTouchCreateFailsInMissingDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "no-such-subdir", "f.txt")
	if err := touch(path, options{}); err == nil {
		t.Fatal("expected an error creating a file in a missing directory")
	}
}

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if New().Synopsis() == "" {
		t.Error("Synopsis() = empty")
	}
}
