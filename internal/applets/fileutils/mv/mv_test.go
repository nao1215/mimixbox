package mv_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mv"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := mv.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRename(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	writeFile(t, src, "hello\n")

	_, errOut, err := run(t, "", src, dest)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, statErr := os.Stat(src); !os.IsNotExist(statErr) {
		t.Errorf("source %s should no longer exist", src)
	}
	got, readErr := os.ReadFile(dest)
	if readErr != nil {
		t.Fatalf("reading dest: %v", readErr)
	}
	if string(got) != "hello\n" {
		t.Errorf("dest content = %q, want %q", string(got), "hello\n")
	}
}

func TestMoveIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, src, "data\n")

	_, errOut, err := run(t, "", src, destDir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	moved := filepath.Join(destDir, "file.txt")
	got, readErr := os.ReadFile(moved)
	if readErr != nil {
		t.Fatalf("reading moved file: %v", readErr)
	}
	if string(got) != "data\n" {
		t.Errorf("moved content = %q, want %q", string(got), "data\n")
	}
	if _, statErr := os.Stat(src); !os.IsNotExist(statErr) {
		t.Errorf("source %s should no longer exist", src)
	}
}

func TestNoClobberDoesNotOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, src, "new\n")
	existing := filepath.Join(destDir, "file.txt")
	writeFile(t, existing, "old\n")

	if _, _, err := run(t, "", "-n", src, destDir); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// -n must keep the existing destination untouched.
	got, readErr := os.ReadFile(existing)
	if readErr != nil {
		t.Fatalf("reading existing file: %v", readErr)
	}
	if string(got) != "old\n" {
		t.Errorf("dest content = %q, want %q (must not overwrite)", string(got), "old\n")
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "mv: missing file operand\n" {
		t.Errorf("stderr = %q, want %q", errOut, "mv: missing file operand\n")
	}
}

func TestMissingDestinationOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "only-src")
	if err == nil {
		t.Fatal("expected error for missing destination operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	want := "mv: missing destination file operand after 'only-src'\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}
