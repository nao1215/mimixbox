package realpath_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/realpath"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := realpath.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunExistingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	want, err := filepath.EvalSymlinks(file)
	if err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, file)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != want+"\n" {
		t.Errorf("out = %q, want %q", out, want+"\n")
	}
}

func TestRunNoSymlinks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	// -s keeps the link path (only cleaning it), but resolves the dir symlink
	// that t.TempDir() may itself sit under.
	cleanDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	wantLink := filepath.Join(cleanDir, "link.txt")

	out, _, err := run(t, "-s", link)
	if err != nil {
		t.Fatalf("Run -s error = %v", err)
	}
	if out != wantLink+"\n" {
		t.Errorf("-s out = %q, want %q", out, wantLink+"\n")
	}

	// Default resolves the symlink to the target.
	wantTarget, err := filepath.EvalSymlinks(link)
	if err != nil {
		t.Fatal(err)
	}
	out, _, err = run(t, link)
	if err != nil {
		t.Fatalf("Run default error = %v", err)
	}
	if out != wantTarget+"\n" {
		t.Errorf("default out = %q, want %q", out, wantTarget+"\n")
	}
}

func TestRunMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cleanDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	nonexistent := filepath.Join(dir, "does", "not", "exist")
	want := filepath.Join(cleanDir, "does", "not", "exist")

	out, _, err := run(t, "-m", nonexistent)
	if err != nil {
		t.Fatalf("Run -m error = %v", err)
	}
	if out != want+"\n" {
		t.Errorf("-m out = %q, want %q", out, want+"\n")
	}
}

func TestRunMissingWithoutFlagErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	nonexistent := filepath.Join(dir, "nope")

	_, errOut, err := run(t, nonexistent)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(errOut, "realpath:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("stderr = %q", errOut)
	}
}
