package mountpoint

import (
	"path/filepath"
	"testing"
)

// TestMountedRoot reports that "/" is a mountpoint: it is its own parent, so the
// inode-equality branch is taken.
func TestMountedRoot(t *testing.T) {
	t.Parallel()
	got, err := mounted("/")
	if err != nil {
		t.Fatalf("mounted(/) err = %v", err)
	}
	if !got {
		t.Error("mounted(/) = false, want true")
	}
}

// TestMountedRegularDir reports that an ordinary directory inside a temp dir is
// not a mountpoint (same device, different inode from its parent).
func TestMountedRegularDir(t *testing.T) {
	t.Parallel()
	got, err := mounted(t.TempDir())
	if err != nil {
		t.Fatalf("mounted(tempdir) err = %v", err)
	}
	if got {
		t.Error("mounted(tempdir) = true, want false")
	}
}

// TestMountedStatError surfaces the stat failure when the directory is missing.
func TestMountedStatError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	got, err := mounted(missing)
	if err == nil {
		t.Fatal("mounted on missing path should error")
	}
	if got {
		t.Error("mounted on error should return false")
	}
}
