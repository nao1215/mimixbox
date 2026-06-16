package ln_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/ln"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: &bytes.Buffer{}, Out: out, Err: errBuf}
	err := ln.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeTarget(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestHardLink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	writeTarget(t, target, "hello\n")

	_, errOut, err := run(t, target, link)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	got, readErr := os.ReadFile(link)
	if readErr != nil {
		t.Fatalf("link not created: %v", readErr)
	}
	if string(got) != "hello\n" {
		t.Errorf("link content = %q, want %q", got, "hello\n")
	}

	// A hard link shares the inode, so writing through one is visible via the
	// other.
	writeTarget(t, target, "changed\n")
	got, _ = os.ReadFile(link)
	if string(got) != "changed\n" {
		t.Errorf("hard link does not share content: got %q", got)
	}
}

func TestSymbolicLink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "sym.txt")
	writeTarget(t, target, "hi\n")

	_, errOut, err := run(t, "-s", target, link)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	info, lerr := os.Lstat(link)
	if lerr != nil {
		t.Fatalf("symlink not created: %v", lerr)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("link mode = %v, want symlink", info.Mode())
	}
}

func TestForceOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	writeTarget(t, target, "data\n")

	// Pre-existing destination.
	writeTarget(t, link, "old\n")

	// Without -f, creating the link over an existing file fails.
	_, _, err := run(t, target, link)
	if err == nil {
		t.Fatal("expected error overwriting existing destination without -f")
	}

	// With -f, the existing destination is removed first and the link wins.
	_, errOut, err := run(t, "-f", target, link)
	if err != nil {
		t.Fatalf("Run -f error = %v (stderr=%q)", err, errOut)
	}
	got, readErr := os.ReadFile(link)
	if readErr != nil {
		t.Fatalf("link not created: %v", readErr)
	}
	if string(got) != "data\n" {
		t.Errorf("link content = %q, want %q", got, "data\n")
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "ln: missing file operand") {
		t.Errorf("stderr = %q, want missing operand message", errOut)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := ln.New()
	if c.Name() != "ln" {
		t.Errorf("Name() = %q, want %q", c.Name(), "ln")
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestVerboseSymbolic exercises the verbose symbolic-link branch of link().
func TestVerboseSymbolic(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "sym.txt")
	writeTarget(t, target, "hi\n")

	out, errOut, err := run(t, "-s", "-v", target, link)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "->") {
		t.Errorf("verbose output = %q, want '->'", out)
	}
}

// TestVerboseHardLink exercises the verbose hard-link branch of link().
func TestVerboseHardLink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "hard.txt")
	writeTarget(t, target, "hi\n")

	out, errOut, err := run(t, "-v", target, link)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "=>") {
		t.Errorf("verbose output = %q, want '=>'", out)
	}
}

// TestMultipleTargetsIntoDirectory covers run()'s "ln TARGET... DIRECTORY" form.
func TestMultipleTargetsIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeTarget(t, a, "A\n")
	writeTarget(t, b, "B\n")
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "-s", a, b, destDir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if _, lerr := os.Lstat(filepath.Join(destDir, name)); lerr != nil {
			t.Errorf("link %s missing: %v", name, lerr)
		}
	}
}

// TestMultipleTargetsNonDirectory checks the "target is not a directory" error
// for three or more operands ending in a regular file.
func TestMultipleTargetsNonDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	dst := filepath.Join(dir, "dst.txt")
	writeTarget(t, a, "A\n")
	writeTarget(t, b, "B\n")
	writeTarget(t, dst, "D\n")

	_, errOut, err := run(t, a, b, dst)
	if err == nil {
		t.Fatal("expected error: final operand is not a directory")
	}
	if !strings.Contains(errOut, "is not a directory") {
		t.Errorf("stderr = %q, want not-a-directory message", errOut)
	}
}

// TestForceRemovesExistingSymlink covers removeExisting via -f over a symlink.
func TestForceRemovesExistingSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")
	writeTarget(t, target, "data\n")
	// Pre-existing symlink at the destination.
	if err := os.Symlink("nowhere", link); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "-f", "-s", target, link)
	if err != nil {
		t.Fatalf("Run -f error = %v (stderr=%q)", err, errOut)
	}
	got, lerr := os.Readlink(link)
	if lerr != nil {
		t.Fatalf("symlink not created: %v", lerr)
	}
	if got != target {
		t.Errorf("symlink points to %q, want %q", got, target)
	}
}

// TestHardLinkMissingTargetReportsReason covers link()'s error branch and the
// GNU-style reason rendering for a missing target.
func TestHardLinkMissingTargetReportsReason(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, filepath.Join(dir, "nope"), filepath.Join(dir, "link"))
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	if !strings.Contains(errOut, "failed to create hard link") {
		t.Errorf("stderr = %q, want hard-link failure message", errOut)
	}
}

func TestSingleOperandLinksInCwd(t *testing.T) {
	dir := t.TempDir()
	// The target lives in a subdirectory so the link, created in cwd with the
	// target's base name, does not collide with the target itself.
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sub, "target.txt")
	writeTarget(t, target, "body\n")

	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, target)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got, readErr := os.ReadFile(filepath.Join(dir, "target.txt"))
	if readErr != nil {
		t.Fatalf("link not created: %v", readErr)
	}
	if string(got) != "body\n" {
		t.Errorf("content = %q, want %q", got, "body\n")
	}
}
