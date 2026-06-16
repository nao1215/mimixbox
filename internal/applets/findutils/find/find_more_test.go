package find_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTypeSymlink covers matchType's symlink (l) branch using a real symlink
// created in a temp tree.
func TestTypeSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	out, _, err := run(t, dir, "-type", "l")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	if len(got) != 1 || got[0] != link {
		t.Errorf("-type l got %v, want the symlink", got)
	}
}

// TestTypeUnknown covers matchType's default branch: an unrecognized type letter
// matches nothing.
func TestTypeUnknown(t *testing.T) {
	t.Parallel()
	root := tree(t)
	out, _, err := run(t, root, "-type", "x")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("-type x out = %q, want empty", out)
	}
}

// TestPathPredicate covers matchOne's -path branch, which matches against the
// whole path with filepath.Match (where '*' does not cross separators).
func TestPathPredicate(t *testing.T) {
	t.Parallel()
	root := tree(t)
	pattern := filepath.Join(root, "sub", "*.log")
	out, _, err := run(t, root, "-path", pattern)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	want := filepath.Join(root, "sub", "b.log")
	if len(got) != 1 || got[0] != want {
		t.Errorf("-path %q got %v, want [%s]", pattern, got, want)
	}
}

// TestTypePipe covers matchType's FIFO (p) branch when the named pipe can be
// created; on platforms without mkfifo support the test is skipped.
func TestTypePipe(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fifo := filepath.Join(dir, "pipe")
	if err := mkfifo(fifo); err != nil {
		t.Skipf("named pipes unsupported: %v", err)
	}
	out, _, err := run(t, dir, "-type", "p")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	got := lines(out)
	if len(got) != 1 || got[0] != fifo {
		t.Errorf("-type p got %v, want the fifo", got)
	}
}
