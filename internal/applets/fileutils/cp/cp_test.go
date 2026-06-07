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
