package rmdir_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/rmdir"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := rmdir.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRemoveEmptyDir(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "empty")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, statErr := os.Stat(dir); !os.IsNotExist(statErr) {
		t.Errorf("directory %q should have been removed", dir)
	}
}

func TestRemoveNonEmptyDirFails(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "full")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, dir)
	if err == nil {
		t.Fatal("expected error removing non-empty directory")
	}
	want := "rmdir: failed to remove '" + dir + "': Directory not empty"
	if !strings.Contains(errOut, want) {
		t.Errorf("stderr = %q, want substring %q", errOut, want)
	}
	if _, statErr := os.Stat(dir); statErr != nil {
		t.Errorf("non-empty directory %q should still exist: %v", dir, statErr)
	}
}

func TestRemoveParents(t *testing.T) {
	// Not parallel: this test chdirs so a relative operand exercises the
	// GNU rmdir -p semantics of removing only the operand's own components.
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "a", "b", "c"), 0o755); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(base); err != nil {
		t.Fatal(err)
	}

	if _, _, rerr := run(t, "-p", filepath.Join("a", "b", "c")); rerr != nil {
		t.Fatalf("Run error = %v", rerr)
	}
	for _, p := range []string{
		filepath.Join("a", "b", "c"),
		filepath.Join("a", "b"),
		"a",
	} {
		if _, statErr := os.Stat(p); !os.IsNotExist(statErr) {
			t.Errorf("directory %q should have been removed", p)
		}
	}
	// The walk stops at the operand's top component; base itself survives.
	if _, statErr := os.Stat(base); statErr != nil {
		t.Errorf("base %q should still exist: %v", base, statErr)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "rmdir: missing operand") {
		t.Errorf("stderr = %q, want rmdir: missing operand", errOut)
	}
}

func TestRunRejectsNonDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "regular.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, file)
	if err == nil {
		t.Fatal("expected error when removing a non-directory")
	}
	if !strings.Contains(errOut, "Not a directory") {
		t.Errorf("stderr = %q, want 'Not a directory'", errOut)
	}
	if _, statErr := os.Stat(file); statErr != nil {
		t.Errorf("file must not be removed by rmdir, stat error = %v", statErr)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	for _, want := range []string{"Usage: rmdir", "Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help missing %q\n%s", want, out)
		}
	}
}
