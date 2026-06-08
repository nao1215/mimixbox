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
