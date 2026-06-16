package cp_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/cp"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := cp.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// runStdin is like run but lets the caller supply stdin, which the interactive
// (-i) overwrite prompt reads its answer from.
func runStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := cp.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestSynopsisAndName(t *testing.T) {
	t.Parallel()
	c := cp.New()
	if c.Name() != "cp" {
		t.Errorf("Name() = %q, want cp", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

func TestRunCopyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	want := []byte("hello copy\n")
	if err := os.WriteFile(src, want, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("dst content = %q, want %q", got, want)
	}
}

func TestRunCopyIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	destDir := filepath.Join(dir, "out")
	want := []byte("into dir\n")
	if err := os.WriteFile(src, want, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, src, destDir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "src.txt"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("copied content = %q, want %q", got, want)
	}
}

func TestRunCopyDirectoryRecursive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "tree")
	inner := filepath.Join(srcDir, "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "b.txt"), []byte("b\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, "-r", srcDir, destDir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	// dest exists, so the tree lands under dest/tree.
	if got, err := os.ReadFile(filepath.Join(destDir, "tree", "a.txt")); err != nil || string(got) != "a\n" {
		t.Errorf("a.txt = %q err = %v", got, err)
	}
	if got, err := os.ReadFile(filepath.Join(destDir, "tree", "inner", "b.txt")); err != nil || string(got) != "b\n" {
		t.Errorf("inner/b.txt = %q err = %v", got, err)
	}
}

func TestRunCopyDirectoryWithoutRecursive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "tree")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, srcDir, destDir)
	if err == nil {
		t.Fatal("expected error copying directory without -r")
	}
	want := "cp: --recursive is not specified: omitting directory: " + srcDir
	if !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want to contain %q", errOut, want)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()

	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "cp: missing file operand") {
		t.Errorf("stderr = %q, want missing file operand", errOut)
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "only.txt")
	if err := os.WriteFile(src, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, errOut, err = run(t, src)
	if err == nil {
		t.Fatal("expected error for missing destination operand")
	}
	if !strings.Contains(errOut, "cp: missing destination file operand after '"+src+"'") {
		t.Errorf("stderr = %q, want missing destination operand", errOut)
	}
}

func TestRunMultipleSourcesRequireDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	dst := filepath.Join(dir, "dst.txt") // a regular file, not a directory
	for _, f := range []string{a, b} {
		if err := os.WriteFile(f, []byte("x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	_, errOut, err := run(t, a, b, dst)
	if err == nil {
		t.Fatal("expected error when copying multiple sources onto a non-directory")
	}
	if !strings.Contains(errOut, "is not a directory") {
		t.Errorf("stderr = %q, want 'is not a directory'", errOut)
	}
	// The copy must be refused before creating dst from the sources.
	if _, statErr := os.Stat(dst); !os.IsNotExist(statErr) {
		t.Errorf("dst should not have been created, stat error = %v", statErr)
	}
}

func TestRunCopyDirIntoItself(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("y\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "-r", src, filepath.Join(src, "child"))
	if err == nil {
		t.Fatal("expected error when copying a directory into its own subtree")
	}
	if !strings.Contains(errOut, "into itself") {
		t.Errorf("stderr = %q, want 'into itself'", errOut)
	}
}

func TestRunCopyFileOntoItselfViaDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	content := []byte("keep me\n")
	if err := os.WriteFile(src, content, 0o600); err != nil {
		t.Fatal(err)
	}

	// "cp dir/a.txt dir" resolves the target to dir/a.txt == src; it must be
	// rejected rather than truncating the source in place.
	_, errOut, err := run(t, src, dir)
	if err == nil {
		t.Fatal("expected error when the resolved target equals the source")
	}
	if !strings.Contains(errOut, "are the same file") {
		t.Errorf("stderr = %q, want 'are the same file'", errOut)
	}
	got, readErr := os.ReadFile(src)
	if readErr != nil || string(got) != string(content) {
		t.Errorf("source was modified: content=%q err=%v", got, readErr)
	}
}

func TestRunPreservesFileMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "script.sh")
	dst := filepath.Join(dir, "copy.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("dst mode = %o, want 755 (execute bit must not be stripped)", info.Mode().Perm())
	}
}

func TestRunPreservesDirMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "private")
	if err := os.Mkdir(src, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "private_copy")
	if _, _, err := run(t, "-r", src, dst); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("dst dir mode = %o, want 700 (a private tree must not be widened)", info.Mode().Perm())
	}
}

func TestRunForceOverwritesReadOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old\n"), 0o400); err != nil {
		t.Fatal(err)
	}

	// Without -f, a read-only destination cannot be opened for writing.
	if _, _, err := run(t, src, dst); err == nil {
		t.Skip("environment allows writing a read-only file (likely running as root); skipping")
	}

	if _, _, err := run(t, "-f", src, dst); err != nil {
		t.Fatalf("cp -f error = %v", err)
	}
	got, _ := os.ReadFile(dst) //nolint:gosec // test-written file
	if string(got) != "new\n" {
		t.Errorf("dst content = %q, want new", got)
	}
}

func TestRunRecursiveDashRAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "tree")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "copy")
	// -R must work the same as -r.
	if _, _, err := run(t, "-R", src, dst); err != nil {
		t.Fatalf("cp -R error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "f.txt")); err != nil {
		t.Errorf("cp -R did not copy the tree: %v", err)
	}
}

func TestRunNoClobber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "-n", src, dst); err != nil {
		t.Fatalf("cp -n error = %v", err)
	}
	got, _ := os.ReadFile(dst) //nolint:gosec // test-written file
	if string(got) != "old\n" {
		t.Errorf("cp -n overwrote the destination: %q", got)
	}
}

func TestRunArchiveImpliesRecursiveAndPreserve(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "tree")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "copy")
	if _, _, err := run(t, "-a", src, dst); err != nil {
		t.Fatalf("cp -a error = %v", err)
	}
	info, err := os.Stat(filepath.Join(dst, "run.sh"))
	if err != nil {
		t.Fatalf("cp -a did not recurse: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("cp -a should preserve mode, got %o", info.Mode().Perm())
	}
}

// TestRunVerboseFile covers the -v reporting branch of cpFile.
func TestRunVerboseFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("v\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "-v", src, dst)
	if err != nil {
		t.Fatalf("cp -v error = %v", err)
	}
	want := "'" + src + "' -> '" + dst + "'"
	if !strings.Contains(out, want) {
		t.Errorf("stdout = %q, want to contain %q", out, want)
	}
}

// TestRunVerboseDir covers the -v reporting branch inside cpDir.
func TestRunVerboseDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "tree")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "copy")
	out, _, err := run(t, "-r", "-v", src, dst)
	if err != nil {
		t.Fatalf("cp -r -v error = %v", err)
	}
	if !strings.Contains(out, "f.txt'") {
		t.Errorf("verbose stdout = %q, want to mention f.txt", out)
	}
}

// TestRunInteractiveYesOverwrites covers question() returning true: an "y"
// answer overwrites the existing destination.
func TestRunInteractiveYesOverwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _, err := runStdin(t, "y\n", "-i", src, dst)
	if err != nil {
		t.Fatalf("cp -i error = %v", err)
	}
	if !strings.Contains(out, "overwrite") {
		t.Errorf("stdout = %q, want overwrite prompt", out)
	}
	got, _ := os.ReadFile(dst) //nolint:gosec // test-written file
	if string(got) != "new\n" {
		t.Errorf("dst content = %q, want new (yes should overwrite)", got)
	}
}

// TestRunInteractiveNoKeeps covers question() returning false: a "n" answer
// leaves the existing destination unchanged.
func TestRunInteractiveNoKeeps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runStdin(t, "n\n", "-i", src, dst); err != nil {
		t.Fatalf("cp -i error = %v", err)
	}
	got, _ := os.ReadFile(dst) //nolint:gosec // test-written file
	if string(got) != "old\n" {
		t.Errorf("dst content = %q, want old (no should not overwrite)", got)
	}
}

// TestRunMissingSource covers the os.Stat error branch in cp().
func TestRunMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")
	dst := filepath.Join(dir, "dst.txt")
	_, errOut, err := run(t, missing, dst)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if !strings.Contains(errOut, "cp:") {
		t.Errorf("stderr = %q, want cp: prefix", errOut)
	}
}

// TestRunNoDereferenceOntoExistingLink covers copySymlink's branch that removes
// an existing destination symlink before recreating it.
func TestRunNoDereferenceOntoExistingLink(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")
	// Pre-create a different symlink at the destination so copySymlink must
	// remove it before writing the new one.
	if err := os.Symlink("somewhere-else", dst); err != nil {
		t.Fatal(err)
	}

	if _, stderr, err := run(t, "-P", link, dst); err != nil {
		t.Fatalf("cp -P error = %v (%s)", err, stderr)
	}
	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatal(err)
	}
	if target != "real.txt" {
		t.Errorf("symlink target = %q, want real.txt (existing link should be replaced)", target)
	}
}

// TestRunNoClobberSkipsExistingLink covers copySymlink's -n early return.
func TestRunNoClobberSkipsExistingLink(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")
	if err := os.Symlink("original-target", dst); err != nil {
		t.Fatal(err)
	}

	if _, stderr, err := run(t, "-P", "-n", link, dst); err != nil {
		t.Fatalf("cp -P -n error = %v (%s)", err, stderr)
	}
	if target, _ := os.Readlink(dst); target != "original-target" {
		t.Errorf("symlink target = %q, want original-target (-n must not replace it)", target)
	}
}

// TestRunSymlinkSameFile covers the early same-path guard in cp() for the
// copy-as-link path: "cp -P dir/link dir" resolves the target to dir/link.
func TestRunSymlinkSameFile(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	_, errOut, err := run(t, "-P", link, dir)
	if err == nil {
		t.Fatal("expected error when the resolved symlink target equals the source")
	}
	if !strings.Contains(errOut, "is same") {
		t.Errorf("stderr = %q, want 'is same'", errOut)
	}
}

// TestRunFollowCmdlineLink covers the derefCmdline (-H) resolution: a
// command-line symlink is followed and its target copied.
func TestRunFollowCmdlineLink(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")
	if _, stderr, err := run(t, "-H", link, dst); err != nil {
		t.Fatalf("cp -H error = %v (%s)", err, stderr)
	}
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("cp -H should follow the command-line link, got a symlink at %q", dst)
	}
}

// TestRunDirSymlinkToDirSkipped covers cpDir's branch that follows a symlink to
// a directory within a tree but does not recurse into it.
func TestRunDirSymlinkToDirSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	realSub := filepath.Join(src, "realdir")
	if err := os.MkdirAll(realSub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realSub, "inside.txt"), []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink pointing at the sibling directory; default deref follows it but
	// must not recurse into the target's contents.
	if err := os.Symlink("realdir", filepath.Join(src, "dlink")); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst")
	if _, stderr, err := run(t, "-r", src, dst); err != nil {
		t.Fatalf("cp -r error = %v (%s)", err, stderr)
	}
	// dst did not exist, so the tree lands directly at dst. The real directory
	// and its file are copied; the followed dir-symlink is not recursed into,
	// so no dlink/inside.txt is created.
	if _, err := os.Stat(filepath.Join(dst, "realdir", "inside.txt")); err != nil {
		t.Errorf("real directory content not copied: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "dlink", "inside.txt")); err == nil {
		t.Errorf("dir-symlink was recursed into; it should be skipped")
	}
}

// TestRunDirNoClobberSkipsExistingInTree covers cpDir's -n branch that skips a
// file already present in the destination tree.
func TestRunDirNoClobberSkipsExistingInTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "f.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Destination already holds src/f.txt with different content.
	dstTree := filepath.Join(dir, "dst", "src")
	if err := os.MkdirAll(dstTree, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dstTree, "f.txt"), []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "dst")
	if _, stderr, err := run(t, "-r", "-n", src, dst); err != nil {
		t.Fatalf("cp -r -n error = %v (%s)", err, stderr)
	}
	got, _ := os.ReadFile(filepath.Join(dstTree, "f.txt")) //nolint:gosec // test-written file
	if string(got) != "old\n" {
		t.Errorf("existing file in tree was overwritten: %q, want old", got)
	}
}

// symlinkFixture creates dir/real.txt and a dir/link -> real.txt symlink and
// returns their paths.
func symlinkFixture(t *testing.T) (dir, realFile, link string) {
	t.Helper()
	dir = t.TempDir()
	realFile = filepath.Join(dir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link = filepath.Join(dir, "link")
	if err := os.Symlink("real.txt", link); err != nil {
		t.Fatal(err)
	}
	return dir, realFile, link
}

func TestRunNoDereferenceCopiesLink(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")

	if _, stderr, err := run(t, "-P", link, dst); err != nil {
		t.Fatalf("cp -P error = %v (%s)", err, stderr)
	}

	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("cp -P should keep %q a symlink", dst)
	}
	if target, _ := os.Readlink(dst); target != "real.txt" {
		t.Errorf("symlink target = %q, want real.txt", target)
	}
}

func TestRunDereferenceCopiesTarget(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")

	if _, stderr, err := run(t, "-L", link, dst); err != nil {
		t.Fatalf("cp -L error = %v (%s)", err, stderr)
	}

	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("cp -L should copy the target, not the link, for %q", dst)
	}
	if got, _ := os.ReadFile(dst); string(got) != "real\n" {
		t.Errorf("dst content = %q, want %q", got, "real\n")
	}
}

func TestRunDefaultFollowsCommandLineLink(t *testing.T) {
	t.Parallel()
	dir, _, link := symlinkFixture(t)
	dst := filepath.Join(dir, "copy")

	if _, stderr, err := run(t, link, dst); err != nil {
		t.Fatalf("cp error = %v (%s)", err, stderr)
	}
	fi, err := os.Lstat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf("default cp should follow the command-line symlink, got a link at %q", dst)
	}
}

func TestRunNoDereferencePreservesLinkInTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "real.txt"), []byte("real\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.txt", filepath.Join(src, "lnk")); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst")

	if _, stderr, err := run(t, "-d", "-r", src, dst); err != nil {
		t.Fatalf("cp -d -r error = %v (%s)", err, stderr)
	}

	fi, err := os.Lstat(filepath.Join(dst, "lnk"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("cp -d should preserve the symlink within the copied tree")
	}
}

func TestRunArchivePreservesLinkInTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "real.txt"), []byte("real\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.txt", filepath.Join(src, "lnk")); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst")

	if _, stderr, err := run(t, "-a", src, dst); err != nil {
		t.Fatalf("cp -a error = %v (%s)", err, stderr)
	}

	fi, err := os.Lstat(filepath.Join(dst, "lnk"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("cp -a should imply -d and preserve the symlink within the tree")
	}
}
