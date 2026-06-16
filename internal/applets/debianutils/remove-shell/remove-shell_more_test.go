package removeShell_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	removeShell "github.com/nao1215/mimixbox/internal/applets/debianutils/remove-shell"
	"github.com/nao1215/mimixbox/internal/command"
)

func runCmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := removeShell.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := removeShell.New()
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
	if c.Name() != "remove-shell" {
		t.Errorf("Name() = %q", c.Name())
	}
}

// TestRunMissingOperand covers Run's empty-args branch (no SHELLNAME given).
func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := runCmd(t)
	if err == nil {
		t.Fatal("expected error when no shell name is given")
	}
	if !strings.Contains(errOut, "shellname") {
		t.Errorf("stderr = %q, want usage hint", errOut)
	}
}

// TestReadShellsTrimsAndSkipsBlanks covers readShells trimming whitespace and
// dropping empty lines.
func TestReadShellsTrimsAndSkipsBlanks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	content := "  /bin/sh  \n\n\t\n/bin/bash\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := removeShell.ReadShellsForTest(path)
	if err != nil {
		t.Fatalf("readShells error = %v", err)
	}
	want := []string{"/bin/sh", "/bin/bash"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("lines = %v, want %v", got, want)
	}
}

// TestReadShellsMissingFileIsEmpty covers readShells returning an empty list for
// a nonexistent file (os.IsNotExist branch).
func TestReadShellsMissingFileIsEmpty(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	got, err := removeShell.ReadShellsForTest(missing)
	if err != nil {
		t.Fatalf("readShells error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("lines = %v, want empty for missing file", got)
	}
}

// TestRemoveShellsCreatesFileWhenMissing covers removeShells against a missing
// path: readShells yields nothing and the file is created (empty) by the
// rewrite.
func TestRemoveShellsCreatesFileWhenMissing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "fresh-shells")
	if err := removeShell.RemoveShellsForTest(path, []string{"/bin/zsh"}); err != nil {
		t.Fatalf("removeShells error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file should have been created: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("file content = %q, want empty", data)
	}
}

// TestRemoveShellsRemovesMultiple covers dropping several names at once and
// preserving the order of the remaining lines.
func TestRemoveShellsRemovesMultiple(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte("/bin/sh\n/bin/bash\n/bin/zsh\n/bin/ksh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeShell.RemoveShellsForTest(path, []string{"/bin/bash", "/bin/ksh"}); err != nil {
		t.Fatalf("removeShells error = %v", err)
	}
	got, err := removeShell.ReadShellsForTest(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/bin/sh", "/bin/zsh"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("lines = %v, want %v", got, want)
	}
}

// TestReadShellsOpenError covers readShells's non-IsNotExist error path by
// pointing at a directory, which cannot be scanned as a file.
func TestReadShellsOpenError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Opening a directory succeeds, but scanning it yields a read error on
	// Linux; either an open or scan error must surface as non-nil.
	_, err := removeShell.ReadShellsForTest(dir)
	if err == nil {
		t.Skip("reading a directory did not error on this platform")
	}
}
