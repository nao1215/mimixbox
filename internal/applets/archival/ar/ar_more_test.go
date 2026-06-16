package ar_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLongMemberNameTruncated drives writeMember's trunc() path: a member name
// longer than the 16-byte ar name field forces the header to be rebuilt with a
// truncated name. The archive must still list and round-trip.
func TestLongMemberNameTruncated(t *testing.T) {
	// Uses t.Chdir, so it cannot be parallel.
	dir := t.TempDir()
	longName := "this_is_a_very_long_member_name.txt" // > 16 bytes
	src := filepath.Join(dir, longName)
	write(t, src, "payload")
	archive := filepath.Join(dir, "long.a")

	if _, errOut, err := run(t, "r", archive, src); err != nil {
		t.Fatalf("create err = %v (stderr=%q)", err, errOut)
	}

	out, _, err := run(t, "t", archive)
	if err != nil {
		t.Fatalf("list err = %v", err)
	}
	// The stored name is the (truncated, slash-trimmed) ar name field; it must be
	// a 15-byte prefix of the original (16 minus the trailing '/').
	listed := strings.TrimSpace(out)
	if listed == "" || !strings.HasPrefix(longName, listed) {
		t.Errorf("listed name %q is not a prefix of %q", listed, longName)
	}
	if len(listed) > 16 {
		t.Errorf("listed name %q longer than 16 bytes", listed)
	}
}

// TestCreateMissingSourceFile covers create()'s os.Stat error branch when a
// named member does not exist.
func TestCreateMissingSourceFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "x.a")
	missing := filepath.Join(dir, "nope.o")

	_, errOut, err := run(t, "r", archive, missing)
	if err == nil {
		t.Fatal("expected error for missing member")
	}
	if !strings.Contains(errOut, "ar:") {
		t.Errorf("stderr = %q, want ar diagnostic", errOut)
	}
}

// TestCreateUnwritableArchive covers create()'s os.Create error branch: writing
// the archive into a path whose parent is not a directory fails.
func TestCreateUnwritableArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	member := filepath.Join(dir, "m.o")
	write(t, member, "data")
	// A regular file used as a directory component makes os.Create fail.
	notDir := filepath.Join(dir, "afile")
	write(t, notDir, "x")
	archive := filepath.Join(notDir, "out.a")

	_, errOut, err := run(t, "r", archive, member)
	if err == nil {
		t.Fatal("expected error creating archive under a non-directory")
	}
	if !strings.Contains(errOut, "ar:") {
		t.Errorf("stderr = %q, want ar diagnostic", errOut)
	}
}

// TestExtractCorruptArchive covers extract()/readArchive error handling on an
// archive whose member header is truncated/corrupt.
func TestExtractCorruptArchive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := filepath.Join(dir, "corrupt.a")
	// Valid magic, then a too-short member header.
	write(t, bad, "!<arch>\nshort-header")

	_, errOut, err := run(t, "x", bad)
	if err == nil {
		t.Fatal("expected error for corrupt archive")
	}
	if !strings.Contains(errOut, "ar:") {
		t.Errorf("stderr = %q, want ar diagnostic", errOut)
	}
}

// TestExtractVerbose covers extract()'s verbose ("x - name") branch.
func TestExtractVerbose(t *testing.T) {
	// Uses t.Chdir, so it cannot be parallel.
	dir := t.TempDir()
	src := filepath.Join(dir, "ev.txt")
	write(t, src, "hello")
	archive := filepath.Join(dir, "ev.a")
	if _, _, err := run(t, "r", archive, src); err != nil {
		t.Fatalf("create err = %v", err)
	}

	extractDir := filepath.Join(dir, "out")
	if err := os.Mkdir(extractDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(extractDir)
	out, _, err := run(t, "xv", archive)
	if err != nil {
		t.Fatalf("extract err = %v", err)
	}
	if !strings.Contains(out, "x - ev.txt") {
		t.Errorf("verbose extract out = %q, want 'x - ev.txt'", out)
	}
}
