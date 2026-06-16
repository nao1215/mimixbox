package uncompress_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKeepFlag covers the -k branch of processFile: the .Z input is decompressed
// but left in place.
func TestKeepFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zf := filepath.Join(dir, "keep.Z")
	if err := os.WriteFile(zf, zbytes(t, "kept"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, nil, "-k", zf); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "keep"))
	if err != nil {
		t.Fatalf("expected decompressed file: %v", err)
	}
	if string(got) != "kept" {
		t.Errorf("decompressed = %q, want kept", got)
	}
	if _, statErr := os.Stat(zf); statErr != nil {
		t.Error("-k should keep the .Z input")
	}
}

// TestExistingOutputBlocksWithoutForce covers the "already exists" guard.
func TestExistingOutputBlocksWithoutForce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zf := filepath.Join(dir, "doc.Z")
	if err := os.WriteFile(zf, zbytes(t, "fresh"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Pre-create the output so the decompress refuses to clobber it.
	out := filepath.Join(dir, "doc")
	if err := os.WriteFile(out, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, nil, zf)
	if err == nil {
		t.Fatal("expected an error when the output already exists")
	}
	if !strings.Contains(errOut, "already exists") {
		t.Errorf("stderr = %q, want an 'already exists' message", errOut)
	}
	// The existing file must be untouched and the .Z kept.
	got, _ := os.ReadFile(out) //nolint:gosec // test-written file
	if string(got) != "old" {
		t.Errorf("output should be untouched, got %q", got)
	}
}

// TestForceOverwrites covers the -f branch: an existing output is overwritten.
func TestForceOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zf := filepath.Join(dir, "doc.Z")
	if err := os.WriteFile(zf, zbytes(t, "brand new"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "doc")
	if err := os.WriteFile(out, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, nil, "-f", zf); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(out) //nolint:gosec // test-written file
	if string(got) != "brand new" {
		t.Errorf("output = %q, want overwritten with 'brand new'", got)
	}
}

// TestMissingInputFile covers the os.Open error branch of processFile.
func TestMissingInputFile(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, filepath.Join(t.TempDir(), "nope.Z"))
	if err == nil {
		t.Fatal("expected an error for a missing input file")
	}
	if !strings.Contains(errOut, "uncompress:") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestMultipleFilesReportEachFailure verifies that one bad operand does not stop
// the others and still sets a non-zero exit.
func TestMultipleFilesReportEachFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.Z")
	if err := os.WriteFile(good, zbytes(t, "ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(dir, "missing.Z")

	_, errOut, err := run(t, nil, bad, good)
	if err == nil {
		t.Fatal("expected a non-zero exit because one operand failed")
	}
	if !strings.Contains(errOut, "uncompress:") {
		t.Errorf("stderr = %q", errOut)
	}
	// The good file was still decompressed.
	got, rerr := os.ReadFile(filepath.Join(dir, "good"))
	if rerr != nil || string(got) != "ok" {
		t.Errorf("good file not decompressed: %v %q", rerr, got)
	}
}
