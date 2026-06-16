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

// TestRelativeSymlink checks that -s -r stores the target relative to the
// link's own directory rather than as the literal (absolute) operand.
func TestRelativeSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// target lives in dir/a/target.txt, link in dir/b/link.txt, so the relative
	// target from the link's directory is "../a/target.txt".
	aDir := filepath.Join(dir, "a")
	bDir := filepath.Join(dir, "b")
	if err := os.MkdirAll(aDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(bDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(aDir, "target.txt")
	link := filepath.Join(bDir, "link.txt")
	writeTarget(t, target, "rel\n")

	_, errOut, err := run(t, "-s", "-r", target, link)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got, lerr := os.Readlink(link)
	if lerr != nil {
		t.Fatalf("symlink not created: %v", lerr)
	}
	want := filepath.Join("..", "a", "target.txt")
	if got != want {
		t.Errorf("symlink target = %q, want relative %q", got, want)
	}
	// The relative symlink must still resolve to the original content.
	content, rerr := os.ReadFile(link)
	if rerr != nil {
		t.Fatalf("relative symlink does not resolve: %v", rerr)
	}
	if string(content) != "rel\n" {
		t.Errorf("resolved content = %q, want %q", content, "rel\n")
	}
}

// TestTargetDirectory checks that -t DIR links each operand into DIR using the
// operand's base name (destination-first form).
func TestTargetDirectory(t *testing.T) {
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

	_, errOut, err := run(t, "-s", "-t", destDir, a, b)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if _, lerr := os.Lstat(filepath.Join(destDir, name)); lerr != nil {
			t.Errorf("link %s missing in target dir: %v", name, lerr)
		}
	}
}

// TestNoTargetDirectoryUsesFileDest checks that -T treats the destination as a
// normal file even when a directory of the same name would otherwise be used.
func TestNoTargetDirectoryUsesFileDest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	writeTarget(t, target, "T\n")
	// A directory named "link" exists; without -T, ln would place the link
	// inside it. With -T the link itself must be created (and fail because the
	// directory already occupies the path).
	linkDir := filepath.Join(dir, "link")
	if err := os.Mkdir(linkDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "-s", "-T", target, linkDir)
	if err == nil {
		t.Fatal("expected error: -T must not descend into an existing directory")
	}
	// The directory must remain a directory (no link created inside it).
	info, serr := os.Lstat(linkDir)
	if serr != nil || !info.IsDir() {
		t.Errorf("link path is no longer the original directory: info=%v err=%v", info, serr)
	}
	if _, lerr := os.Lstat(filepath.Join(linkDir, "target.txt")); lerr == nil {
		t.Error("-T must not create a link inside the directory")
	}
}

// TestNoTargetDirectoryRejectsExtraOperand checks -T rejects more than two
// operands.
func TestNoTargetDirectoryRejectsExtraOperand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeTarget(t, a, "A\n")
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, "-s", "-T", a, b, destDir)
	if err == nil {
		t.Fatal("expected error for extra operand with -T")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want extra-operand message", errOut)
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
