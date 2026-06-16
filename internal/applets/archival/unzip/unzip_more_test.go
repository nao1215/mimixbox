package unzip_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractDirectoryEntry covers the IsDir branch of extractFile: a zip entry
// whose name ends in "/" is recreated as a directory.
func TestExtractDirectoryEntry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "withdir.zip")

	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// A directory header (trailing slash). Give it a real mode so MkdirAll
	// produces a writable directory the nested file can then be created in.
	dirHdr := &zip.FileHeader{Name: "emptydir/"}
	dirHdr.SetMode(0o755 | os.ModeDir)
	if _, err := zw.CreateHeader(dirHdr); err != nil {
		t.Fatal(err)
	}
	w, err := zw.Create("emptydir/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("inside")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	dest := filepath.Join(dir, "out")
	if _, errOut, err := run(t, "-d", dest, archive); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}

	info, err := os.Stat(filepath.Join(dest, "emptydir"))
	if err != nil {
		t.Fatalf("directory entry not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("emptydir should be a directory")
	}
	got, err := os.ReadFile(filepath.Join(dest, "emptydir", "file.txt"))
	if err != nil || string(got) != "inside" {
		t.Errorf("nested file not extracted: %v %q", err, got)
	}
}

// TestVerboseReportsEachFile covers the verbose branch of extractFile, which
// prints each extracted entry to stderr.
func TestVerboseReportsEachFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "v.zip")
	makeZip(t, archive, map[string]string{"note.txt": "hi"})

	dest := filepath.Join(dir, "out")
	_, errOut, err := run(t, "-v", "-d", dest, archive)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(errOut, "extracting:") || !strings.Contains(errOut, "note.txt") {
		t.Errorf("verbose stderr = %q, want an 'extracting: note.txt' line", errOut)
	}
}

// TestZipSlipRejected covers the safeJoin guard that rejects entries whose names
// try to escape the destination directory (a path-traversal "zip slip").
func TestZipSlipRejected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "evil.zip")

	f, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// Use a raw header so the malicious "../" name is preserved verbatim.
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "../escape.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("pwned")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	dest := filepath.Join(dir, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, "-d", dest, archive)
	if err == nil {
		t.Fatal("expected the path-traversal entry to be rejected")
	}
	if !strings.Contains(errOut, "outside the destination directory") {
		t.Errorf("stderr = %q, want a zip-slip rejection", errOut)
	}
	// The file must not have been written outside dest.
	if _, statErr := os.Stat(filepath.Join(dir, "escape.txt")); statErr == nil {
		t.Error("zip-slip wrote a file outside the destination directory")
	}
}
