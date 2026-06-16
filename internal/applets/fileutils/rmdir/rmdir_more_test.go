package rmdir_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/rmdir"
)

// TestSynopsis covers the one-line description helper.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if got := rmdir.New().Synopsis(); got != "Remove directory" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestMultipleFailuresKeepsFirstError drives keep(): when two operands both
// fail, a single failure error is still returned (the first one is kept).
func TestMultipleFailuresKeepsFirstError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	nonEmptyA := filepath.Join(dir, "a")
	nonEmptyB := filepath.Join(dir, "b")
	for _, d := range []string{nonEmptyA, nonEmptyB} {
		if err := os.MkdirAll(filepath.Join(d, "child"), 0o755); err != nil {
			t.Fatalf("setup %s: %v", d, err)
		}
	}

	_, errOut, err := run(t, nonEmptyA, nonEmptyB)
	if err == nil {
		t.Fatal("expected a failure when both directories are non-empty")
	}
	// Both failures are reported on stderr even though one error is returned.
	for _, name := range []string{"'" + nonEmptyA + "'", "'" + nonEmptyB + "'"} {
		if !strings.Contains(errOut, name) {
			t.Errorf("stderr = %q, want a diagnostic for %s", errOut, name)
		}
	}
}

// TestParentsStopsAtFirstFailure verifies remove() aborts the ancestor walk when
// a parent is not empty: 'rmdir -p a/b/c' fails at b when b also holds a sibling.
func TestParentsStopsAtFirstFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// a/b/c is empty, but b also contains a sibling so b cannot be removed.
	abc := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(abc, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	sibling := filepath.Join(root, "a", "b", "sibling")
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Fatalf("mkdir sibling: %v", err)
	}

	_, _, err := run(t, "-p", abc)
	if err == nil {
		t.Fatal("expected failure: ancestor b is not empty")
	}
	// c was removed, but b survives because of the sibling.
	if _, serr := os.Stat(abc); serr == nil {
		t.Errorf("leaf %s should have been removed", abc)
	}
	if _, serr := os.Stat(filepath.Join(root, "a", "b")); serr != nil {
		t.Errorf("non-empty ancestor b should survive: %v", serr)
	}
}

// TestParentsRemovesAncestors verifies the success path of -p where every
// ancestor of a relative operand becomes empty and is removed in turn. Running
// from inside the temp root with a relative path "x/y/z" lets the ancestor walk
// stop cleanly at "." instead of climbing into system directories.
func TestParentsRemovesAncestors(t *testing.T) {
	root := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if err := os.MkdirAll(filepath.Join("x", "y", "z"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if _, _, err := run(t, "-p", filepath.Join("x", "y", "z")); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	for _, d := range []string{"x/y/z", "x/y", "x"} {
		if _, serr := os.Stat(filepath.Join(root, filepath.FromSlash(d))); serr == nil {
			t.Errorf("ancestor %s should have been removed", d)
		}
	}
	// The temp root itself survives: the walk stops at "." before reaching it.
	if _, serr := os.Stat(root); serr != nil {
		t.Errorf("root should be untouched: %v", serr)
	}
}

// TestVerboseReportsRemoval exercises the -v diagnostic on stdout.
func TestVerboseReportsRemoval(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	if err := os.Mkdir(empty, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	out, _, err := run(t, "-v", empty)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "removing directory") || !strings.Contains(out, empty) {
		t.Errorf("verbose stdout = %q", out)
	}
}
