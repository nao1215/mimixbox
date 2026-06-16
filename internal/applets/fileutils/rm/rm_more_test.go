package rm_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/rm"
)

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := rm.New()
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
	if c.Name() != "rm" {
		t.Errorf("Name() = %q", c.Name())
	}
}

// TestRemoveEmptyDirWithDirFlag covers the -d branch: an empty directory may be
// removed without -r.
func TestRemoveEmptyDirWithDirFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "empty")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, nil, "-d", sub); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if exists(sub) {
		t.Errorf("empty directory %q should have been removed with -d", sub)
	}
}

// TestRemoveNonEmptyDirWithDirFlagErrors covers os.Remove failing on a
// non-empty directory when -d (not -r) is used.
func TestRemoveNonEmptyDirWithDirFlagErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "full")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "inner.txt"))

	_, errOut, err := run(t, nil, "-d", sub)
	if err == nil {
		t.Fatal("expected error removing a non-empty directory with -d")
	}
	if !exists(sub) {
		t.Errorf("directory %q should remain", sub)
	}
	if !strings.Contains(errOut, "rm:") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestVerboseReportsRemoval covers report(): -v prints a removal notice on
// stdout.
func TestVerboseReportsRemoval(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	writeFile(t, f)

	out, _, err := run(t, nil, "-v", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "removed '"+f+"'") {
		t.Errorf("verbose out = %q, want a removal notice", out)
	}
}

// TestVerboseRecursiveReportsDir covers report() on the recursive-directory
// path.
func TestVerboseRecursiveReportsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "inner.txt"))

	out, _, err := run(t, nil, "-rv", sub)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, "removed '"+sub+"'") {
		t.Errorf("verbose out = %q", out)
	}
	if exists(sub) {
		t.Errorf("directory %q should have been removed", sub)
	}
}

// TestInteractiveEOFKeepsFile covers confirm()'s EOF branch: with -i and an
// empty (EOF) answer, the file is kept.
func TestInteractiveEOFKeepsFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	writeFile(t, f)

	// Empty reader => ReadString hits EOF immediately with an empty answer.
	_, errOut, err := run(t, strings.NewReader(""), "-i", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !exists(f) {
		t.Errorf("file %q should remain after EOF answer", f)
	}
	if !strings.Contains(errOut, "remove '"+f+"'?") {
		t.Errorf("prompt = %q", errOut)
	}
}

// TestInteractiveNoDirKeepsDir covers confirm() returning false on the directory
// path (-r -i answered "no").
func TestInteractiveNoDirKeepsDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, strings.NewReader("n\n"), "-r", "-i", sub); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !exists(sub) {
		t.Errorf("directory %q should remain after answering no", sub)
	}
}

// TestForceWithNoOperands covers Run's early return when -f is set and there are
// no operands (no error, no output).
func TestForceWithNoOperands(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, nil, "-f")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" || errOut != "" {
		t.Errorf("expected silent success, got out=%q err=%q", out, errOut)
	}
}

// TestRecursiveAliasUppercaseR covers the -R alias for -r.
func TestRecursiveAliasUppercaseR(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "inner.txt"))

	if _, _, err := run(t, nil, "-R", sub); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if exists(sub) {
		t.Errorf("directory %q should have been removed with -R", sub)
	}
}
