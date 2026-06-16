package shred_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNonexistentFileReported drives shredFile's stat-error path: a missing file
// is reported and yields a failure.
func TestNonexistentFileReported(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "/no/such/file/here")
	if err == nil {
		t.Fatal("expected error for a nonexistent file")
	}
	if !strings.Contains(errOut, "shred: /no/such/file/here") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestEmptyFileOverwrite drives overwrite() with size 0: io.CopyN returns EOF
// immediately, which must be treated as success.
func TestEmptyFileOverwrite(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "")
	if _, _, err := run(t, "-n", "2", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("empty file size = %d, want 0", info.Size())
	}
}

// TestRemoveZeroPassCombined exercises -z together with -u: the file is
// overwritten with zeros, then truncated and removed.
func TestRemoveZeroPassCombined(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "topsecret")
	if _, _, err := run(t, "-z", "-u", "-n", "1", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("file should have been removed after -z -u")
	}
}

// TestMultipleFilesOneMissing verifies that a per-file error does not abort the
// whole run: the good file is still shredded, and the run fails overall.
func TestMultipleFilesOneMissing(t *testing.T) {
	t.Parallel()
	good := tmpFile(t, "keepme")
	missing := filepath.Join(t.TempDir(), "gone")

	_, errOut, err := run(t, missing, good)
	if err == nil {
		t.Fatal("expected an overall failure when one file is missing")
	}
	if !strings.Contains(errOut, "gone") {
		t.Errorf("stderr = %q, want a diagnostic for the missing file", errOut)
	}
	// The good file was still processed: it survives (no -u) at its size.
	info, serr := os.Stat(good)
	if serr != nil {
		t.Fatalf("good file stat: %v", serr)
	}
	if info.Size() != int64(len("keepme")) {
		t.Errorf("good file size = %d, want %d", info.Size(), len("keepme"))
	}
}
