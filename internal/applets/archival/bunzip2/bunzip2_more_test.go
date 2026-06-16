package bunzip2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTbz2SuffixMapping covers outputName's .tbz2 -> .tar mapping and a full
// decompress-to-file run through that suffix.
func TestTbz2SuffixMapping(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "archive.tbz2")
	writeBz2(t, src)

	if _, errOut, err := run(t, nil, src); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "archive.tar"))
	if err != nil {
		t.Fatalf("expected archive.tar: %v", err)
	}
	if string(got) != bz2Text {
		t.Errorf("decompressed = %q", got)
	}
}

// TestTbzSuffixMapping covers outputName's .tbz -> .tar mapping.
func TestTbzSuffixMapping(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "data.tbz")
	writeBz2(t, src)

	if _, errOut, err := run(t, nil, src); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, "data.tar")); err != nil {
		t.Errorf("expected data.tar: %v", err)
	}
}

// TestStdoutMissingFile covers processFile's os.Open error branch on the -c
// path when the named file does not exist.
func TestStdoutMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.bz2")

	_, errOut, err := run(t, nil, "-c", missing)
	if err == nil {
		t.Fatal("expected error for missing file with -c")
	}
	if !strings.Contains(errOut, "bunzip2:") {
		t.Errorf("stderr = %q, want bunzip2 diagnostic", errOut)
	}
}

// TestReplaceMissingFile covers processFile's os.Open error branch on the
// replace path (no -c) when the file does not exist.
func TestReplaceMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "gone.bz2")

	_, errOut, err := run(t, nil, missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(errOut, "bunzip2:") {
		t.Errorf("stderr = %q, want bunzip2 diagnostic", errOut)
	}
}

// TestCorruptData covers decompressStream's io.Copy error branch on bytes that
// are not valid bzip2.
func TestCorruptData(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, []byte("this is not bzip2 data at all"))
	if err == nil {
		t.Fatal("expected error decompressing non-bzip2 input")
	}
	if !strings.Contains(errOut, "bunzip2:") {
		t.Errorf("stderr = %q, want bunzip2 diagnostic", errOut)
	}
}

// TestMultipleFilesPartialFailure covers Run's per-file loop where one file
// succeeds and another (unknown suffix) fails, so failed is set but the good
// file is still decompressed.
func TestMultipleFilesPartialFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "ok.bz2")
	bad := filepath.Join(dir, "noext")
	writeBz2(t, good)
	writeBz2(t, bad)

	_, errOut, err := run(t, nil, good, bad)
	if err == nil {
		t.Fatal("expected error from the bad-suffix file")
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q, want unknown-suffix diagnostic", errOut)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "ok")); statErr != nil {
		t.Errorf("good file should still be decompressed: %v", statErr)
	}
}
