package rm_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/rm"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, in io.Reader, args ...string) (string, string, error) {
	t.Helper()
	if in == nil {
		in = strings.NewReader("")
	}
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	stdio := command.IO{In: in, Out: out, Err: errBuf}
	err := rm.New().Run(context.Background(), stdio, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestRemoveFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	writeFile(t, f)

	_, _, err := run(t, nil, f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if exists(f) {
		t.Errorf("file %q should have been removed", f)
	}
}

func TestRemoveDirWithoutRecursiveErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, nil, sub)
	if err == nil {
		t.Fatal("expected error removing a directory without -r")
	}
	if !exists(sub) {
		t.Errorf("directory %q should remain", sub)
	}
	want := "rm: can't remove " + sub + ": It's directory\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestRemoveDirRecursive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "inner.txt"))

	_, _, err := run(t, nil, "-r", sub)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if exists(sub) {
		t.Errorf("directory %q should have been removed", sub)
	}
}

func TestForceMissingFileSucceedsSilently(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no_such_file.txt")

	out, errOut, err := run(t, nil, "-f", missing)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "" || errOut != "" {
		t.Errorf("expected no output, got out=%q err=%q", out, errOut)
	}
}

func TestMissingFileErrors(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "no_such_file.txt")

	_, errOut, err := run(t, nil, missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	want := "rm: can't remove " + missing + ": No such file or directory exists\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestInteractiveYesRemoves(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	writeFile(t, f)

	_, errOut, err := run(t, strings.NewReader("y\n"), "-i", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if exists(f) {
		t.Errorf("file %q should have been removed after answering yes", f)
	}
	if !strings.Contains(errOut, "remove '"+f+"'?") {
		t.Errorf("prompt = %q, want it to ask about %q", errOut, f)
	}
}

func TestInteractiveNoKeeps(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "a.txt")
	writeFile(t, f)

	_, _, err := run(t, strings.NewReader("n\n"), "-i", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !exists(f) {
		t.Errorf("file %q should remain after answering no", f)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "rm: missing operand") {
		t.Errorf("stderr = %q, want missing operand", errOut)
	}
}

func TestMissingMiddleFileContinues(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	c := filepath.Join(dir, "c.txt")
	missing := filepath.Join(dir, "b.txt")
	writeFile(t, a)
	writeFile(t, c)

	_, _, err := run(t, nil, a, missing, c)
	if err == nil {
		t.Fatal("expected failure because one file is missing")
	}
	if exists(a) || exists(c) {
		t.Errorf("a and c should be removed despite the missing middle file")
	}
}
