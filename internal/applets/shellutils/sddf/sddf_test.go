package sddf_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/sddf"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes sddf with the given stdin and arguments, returning stdout,
// stderr and the error.
func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := sddf.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestFindDuplicates checks the pure detection core: byte-identical files are
// grouped together and unique files are not reported.
func TestFindDuplicates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	dupA := filepath.Join(dir, "a.txt")
	dupB := filepath.Join(dir, "b.txt")
	dupC := filepath.Join(dir, "c.txt")
	uniq := filepath.Join(dir, "unique.txt")

	writeFile(t, dupA, []byte("same content"))
	writeFile(t, dupB, []byte("same content"))
	writeFile(t, dupC, []byte("same content"))
	writeFile(t, uniq, []byte("different"))

	groups := sddf.FindDuplicatesForTest([]string{dupA, dupB, dupC, uniq})

	if len(groups) != 1 {
		t.Fatalf("got %d duplicate groups, want 1: %v", len(groups), groups)
	}
	got := append([]string{}, groups[0]...)
	sort.Strings(got)
	want := []string{dupA, dupB, dupC}
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("group = %v, want %v", got, want)
	}
}

// TestFindDuplicatesNoDuplicates verifies that a set of unique files yields no
// groups.
func TestFindDuplicatesNoDuplicates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeFile(t, a, []byte("aaa"))
	writeFile(t, b, []byte("bbb"))

	groups := sddf.FindDuplicatesForTest([]string{a, b})
	if len(groups) != 0 {
		t.Fatalf("got %d groups, want 0: %v", len(groups), groups)
	}
}

// TestDeleteRemovesDuplicatesKeepsOne runs the full command with --delete and a
// "y" answer on stdin, asserting that duplicates are removed and exactly one
// copy survives.
func TestDeleteRemovesDuplicatesKeepsOne(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	writeFile(t, a, []byte("dup"))
	writeFile(t, b, []byte("dup"))
	writeFile(t, c, []byte("dup"))

	// One "y" per deletion (two duplicates of a three-file group).
	_, _, err := run(t, "y\ny\n", "--delete", "--interactive", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	remaining := 0
	for _, p := range []string{a, b, c} {
		if _, err := os.Stat(p); err == nil {
			remaining++
		}
	}
	if remaining != 1 {
		t.Errorf("remaining files = %d, want 1", remaining)
	}
}

// TestDryRunDeletesNothing asserts that --dry-run reports but never removes.
func TestDryRunDeletesNothing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeFile(t, a, []byte("dup"))
	writeFile(t, b, []byte("dup"))

	out, _, err := run(t, "", "--dry-run", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(a); err != nil {
		t.Errorf("file a was removed during dry-run: %v", err)
	}
	if _, err := os.Stat(b); err != nil {
		t.Errorf("file b was removed during dry-run: %v", err)
	}
	if !strings.Contains(out, "Delete(DryRun)") {
		t.Errorf("dry-run output = %q, want a DryRun line", out)
	}
}

// TestMissingOperand asserts a usage error when no directory is given.
func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "missing directory operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

// TestHelpAndVersion exercises the standard --help / --version flags provided by
// the framework.
func TestHelpAndVersion(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	for _, want := range []string{"Usage: sddf", "Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help out missing %q\n%s", want, out)
		}
	}

	out, _, err = run(t, "", "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "sddf (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}
