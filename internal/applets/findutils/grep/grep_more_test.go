package grep_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWithNameSingleFile covers showName's withName branch: -H forces the file
// name prefix even for a single file (which normally omits it).
func TestWithNameSingleFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "only.txt")
	if err := os.WriteFile(f, []byte("hit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-H", "hit", f)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != f+":hit\n" {
		t.Errorf("-H out = %q, want %q", out, f+":hit\n")
	}
}

// TestNoNameMultipleFiles covers showName's noName branch: -h suppresses the
// file-name prefix that multiple files would otherwise add.
func TestNoNameMultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("hit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("hit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-h", "hit", a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "hit\nhit\n" {
		t.Errorf("-h out = %q, want plain lines without file names", out)
	}
}

// TestCountWithName covers searchFile's "count with name prefix" branch, where
// -c against multiple files prefixes each count with its file name.
func TestCountWithName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(a, []byte("hit\nhit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("hit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", "-c", "hit", a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, a+":2") || !strings.Contains(out, b+":1") {
		t.Errorf("-c out = %q, want per-file counts with name prefix", out)
	}
}

// TestRecursiveUnreadablePath covers walk's error callback: an unreadable
// directory entry makes WalkDir report an error, which grep surfaces and turns
// into exit status 2.
func TestRecursiveUnreadablePath(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read directories regardless of permission bits")
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "locked")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "x.txt"), []byte("hit\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sub, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sub, 0o755) })

	_, errOut, err := run(t, "", "-r", "hit", dir)
	if code := exitCode(t, err); code != 2 {
		t.Errorf("exit = %d, want 2 for an unreadable directory", code)
	}
	if !strings.Contains(errOut, "grep:") {
		t.Errorf("stderr = %q, want a grep error", errOut)
	}
}
