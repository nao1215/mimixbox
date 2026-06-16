package split_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestByBytesKiloSuffix drives parseSize's "K" suffix branch: -b 1K splits into
// 1024-byte pieces.
func TestByBytesKiloSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "k-")
	content := strings.Repeat("a", 1024) + strings.Repeat("b", 512)
	if _, _, err := run(t, content, "-b", "1K", "-", prefix); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, prefix+"aa", strings.Repeat("a", 1024))
	checkFile(t, prefix+"ab", strings.Repeat("b", 512))
}

// TestByBytesMegaSuffix drives parseSize's "M" suffix branch.
func TestByBytesMegaSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "m-")
	// One full MiB plus a small remainder, to produce two files.
	content := strings.Repeat("x", 1024*1024) + "tail"
	if _, _, err := run(t, content, "-b", "1M", "-", prefix); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(prefix + "aa")
	if err != nil {
		t.Fatalf("first piece missing: %v", err)
	}
	if info.Size() != 1024*1024 {
		t.Errorf("first piece size = %d, want %d", info.Size(), 1024*1024)
	}
	checkFile(t, prefix+"ab", "tail")
}

// TestParseSizeRejectsZero verifies parseSize rejects a non-positive count even
// with a suffix.
func TestParseSizeRejectsZero(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "data", "-b", "0K")
	if err == nil {
		t.Fatal("expected error for -b 0K")
	}
	if !strings.Contains(errOut, "invalid number of bytes") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestByLinesExactMultiple drives byLines where the line count is an exact
// multiple of the per-file size: the final file is closed mid-loop and no empty
// trailing file is produced.
func TestByLinesExactMultiple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "exact-")
	if _, _, err := run(t, "1\n2\n3\n4\n", "-l", "2", "-", prefix); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, prefix+"aa", "1\n2\n")
	checkFile(t, prefix+"ab", "3\n4\n")
	// No third file should exist.
	if _, err := os.Stat(prefix + "ac"); !os.IsNotExist(err) {
		t.Errorf("unexpected extra file %sac", prefix)
	}
}

// TestByLinesSingleFile drives byLines where all lines fit in one output file.
func TestByLinesSingleFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	prefix := filepath.Join(dir, "one-")
	if _, _, err := run(t, "alpha\nbeta\n", "-l", "10", "-", prefix); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, prefix+"aa", "alpha\nbeta\n")
}

// TestByLinesDefaultPrefix verifies the default prefix "x" is used when none is
// given. Output files are created in the current directory, so the test runs
// from a temp dir.
func TestByLinesDefaultPrefix(t *testing.T) {
	work := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if _, _, err := run(t, "one\ntwo\n", "-l", "1"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	checkFile(t, filepath.Join(work, "xaa"), "one\n")
	checkFile(t, filepath.Join(work, "xab"), "two\n")
}
