package vi_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestRunReadErrorOnDirectory covers the non-IsNotExist read-error branch of
// Run: passing a directory makes os.ReadFile fail with a non "not exist" error,
// which must be reported and exit non-zero.
func TestRunReadErrorOnDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, ":q\r", dir)
	if err == nil {
		t.Fatal("expected an error when the operand is a directory")
	}
	if !strings.Contains(errOut, "vi:") {
		t.Errorf("stderr = %q, want a vi error prefix", errOut)
	}
}

// TestRunSaveErrorUnwritablePath covers the write-error branch of Run: saving
// into a path whose parent directory does not exist fails on os.WriteFile.
func TestRunSaveErrorUnwritablePath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := filepath.Join(dir, "missing-dir", "file.txt")
	_, errOut, err := run(t, "iX\x1b:wq\r", bad)
	if err == nil {
		t.Fatal("expected a write error for a nonexistent parent directory")
	}
	if !strings.Contains(errOut, "vi:") {
		t.Errorf("stderr = %q, want a vi error prefix", errOut)
	}
}
