package patch_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMalformedHeader covers parseUnified's "--- " without a following "+++ ".
func TestMalformedHeader(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "--- only-one-header\nsome body\n")
	if err == nil {
		t.Fatal("expected error for malformed header")
	}
	if !strings.Contains(errOut, "malformed header") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestHunkBeforeHeader covers parseUnified's "hunk before file header" branch.
func TestHunkBeforeHeader(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "@@ -1,1 +1,1 @@\n a\n")
	if err == nil {
		t.Fatal("expected error for hunk before header")
	}
	if !strings.Contains(errOut, "hunk before file header") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestBadHunkHeader covers parseHunkHeader rejecting a malformed @@ line.
func TestBadHunkHeader(t *testing.T) {
	t.Parallel()
	// Too few fields after @@ to form a valid header.
	diff := "--- f\n+++ f\n@@ -1,1\n a\n"
	_, errOut, err := run(t, diff)
	if err == nil {
		t.Fatal("expected error for bad hunk header")
	}
	if !strings.Contains(errOut, "bad hunk header") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestBadRangeLength covers parseRange's non-numeric length.
func TestBadRangeLength(t *testing.T) {
	t.Parallel()
	diff := "--- f\n+++ f\n@@ -1,x +1,1 @@\n a\n"
	_, errOut, err := run(t, diff)
	if err == nil {
		t.Fatal("expected error for bad range length")
	}
	if !strings.Contains(errOut, "patch:") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestHunkStartBeyondEnd covers applyHunks rejecting a hunk that starts past the
// end of the file.
func TestHunkStartBeyondEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// File has a single line, but the hunk claims to start at line 5.
	diff := "--- " + target + "\n+++ " + target + "\n@@ -5,1 +5,1 @@\n a\n"
	_, errOut, err := run(t, diff)
	if err == nil {
		t.Fatal("expected error for hunk start beyond end")
	}
	if !strings.Contains(errOut, "beyond end of file") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestNoNewlineMarkerIgnored covers parseHunk's handling of the
// "\ No newline at end of file" marker: it is skipped and the patch still
// applies.
func TestNoNewlineMarkerIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	diff := "--- " + target + "\n+++ " + target + "\n" +
		"@@ -1,1 +1,1 @@\n" +
		"-old\n" +
		"\\ No newline at end of file\n" +
		"+new\n" +
		"\\ No newline at end of file\n"
	if _, errOut, err := run(t, diff); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "new\n" {
		t.Errorf("patched = %q, want new", got)
	}
}

// TestTrailingTextEndsHunkBody covers parseHunk's default branch where a line
// that is not part of the hunk body terminates it, and parsing continues. Here
// a non-diff "garbage" line follows the hunk body.
func TestTrailingTextEndsHunkBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	diff := "--- " + target + "\n+++ " + target + "\n" +
		"@@ -1,1 +1,1 @@\n" +
		" one\n" +
		"trailing prose that is not part of the hunk\n"
	if _, errOut, err := run(t, diff); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "one\ntwo\n" {
		t.Errorf("patched = %q", got)
	}
}

// TestEmptyContextLineInHunk covers parseHunk's handling of a truly empty line
// inside a hunk body (treated as a context line containing "").
func TestEmptyContextLineInHunk(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	// Original content: "a", "", "b".
	if err := os.WriteFile(target, []byte("a\n\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The middle empty line is represented as an empty body line in the diff.
	diff := "--- " + target + "\n+++ " + target + "\n" +
		"@@ -1,3 +1,3 @@\n" +
		" a\n" +
		"\n" +
		"-b\n" +
		"+B\n"
	if _, errOut, err := run(t, diff); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "a\n\nB\n" {
		t.Errorf("patched = %q, want a\\n\\nB", got)
	}
}

// TestApplyToMissingTarget covers applyFile's open-error path.
func TestApplyToMissingTarget(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no-such-target.txt")
	diff := "--- " + missing + "\n+++ " + missing + "\n@@ -1,1 +1,1 @@\n a\n"
	_, errOut, err := run(t, diff)
	if err == nil {
		t.Fatal("expected error for missing target file")
	}
	if !strings.Contains(errOut, "cannot open") {
		t.Errorf("stderr = %q", errOut)
	}
}
