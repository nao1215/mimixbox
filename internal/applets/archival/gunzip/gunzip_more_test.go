package gunzip_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTgzSuffix covers outputName's .tgz branch, which becomes ".tar".
func TestTgzSuffix(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "bundle.tgz")
	writeGz(t, src, "payload")

	if _, errOut, err := run(t, nil, src); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "bundle.tar"))
	if err != nil {
		t.Fatalf("expected bundle.tar: %v", err)
	}
	if string(got) != "payload" {
		t.Errorf("decompressed = %q, want payload", got)
	}
}

// TestStdoutMissingFile covers processFile's -c open-error branch.
func TestStdoutMissingFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, "-c", "/no/such/file.gz")
	if err == nil {
		t.Fatal("expected error opening a missing file with -c")
	}
	if !strings.Contains(errOut, "gunzip:") {
		t.Errorf("stderr = %q, want gunzip error", errOut)
	}
}

// TestMissingInputFile covers processFile's (non -c) open-error branch.
func TestMissingInputFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, "/no/such/file.gz")
	if err == nil {
		t.Fatal("expected error opening a missing file")
	}
	if !strings.Contains(errOut, "gunzip:") {
		t.Errorf("stderr = %q, want gunzip error", errOut)
	}
}

// TestBadGzipFileLeavesNoOutput covers decompressStream's NewReader error path
// when invoked through a file (not stdin): a corrupt .gz must fail and report.
func TestBadGzipFileLeavesNoOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "corrupt.gz")
	if err := os.WriteFile(src, []byte("this is not gzip data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, nil, src)
	if err == nil {
		t.Fatal("expected error for a corrupt .gz file")
	}
	if !strings.Contains(errOut, "gunzip:") {
		t.Errorf("stderr = %q, want gunzip error", errOut)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "corrupt")); !os.IsNotExist(statErr) {
		t.Errorf("a partial output file was left behind after corrupt input: %v", statErr)
	}
}

// TestMultipleFilesOneBad covers Run's failed-loop branch: a good file is still
// decompressed while a bad one is reported and the overall run fails.
func TestMultipleFilesOneBad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.gz")
	writeGz(t, good, "ok")
	bad := filepath.Join(dir, "bad") // no recognized suffix

	writeGz(t, bad, "x")

	_, errOut, err := run(t, nil, good, bad)
	if err == nil {
		t.Fatal("expected failure when one file has an unknown suffix")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "good")); statErr != nil {
		t.Errorf("good file should still be decompressed: %v", statErr)
	}
	if !strings.Contains(errOut, "unknown suffix") {
		t.Errorf("stderr = %q, want unknown-suffix message", errOut)
	}
}
