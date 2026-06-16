package unix2dos_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/unix2dos"
)

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if unix2dos.New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestNoArgsIsNoop verifies that with no FILE operands the command does nothing
// and exits successfully (the for-loop body never runs).
func TestNoArgsIsNoop(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err != nil {
		t.Fatalf("Run with no args error = %v", err)
	}
	if out != "" || errOut != "" {
		t.Errorf("no-args run produced output: out=%q err=%q", out, errOut)
	}
}

// TestWriteErrorInReadOnlyDirectory covers the write-error branch of Run.
// ListToFile writes via a temp file in the file's parent directory and renames
// it into place, so a read-only directory (not a read-only file) is what makes
// the write fail. Skipped as root, which bypasses directory permission bits.
func TestWriteErrorInReadOnlyDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission bits; cannot provoke a write error this way")
	}
	if runtime.GOOS == "windows" {
		t.Skip("read-only permission semantics differ on Windows")
	}
	t.Parallel()
	parent := t.TempDir()
	roDir := filepath.Join(parent, "ro")
	if err := os.Mkdir(roDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(roDir, "doc.txt")
	if err := os.WriteFile(file, []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make the directory read-only so the temp file cannot be created there.
	if err := os.Chmod(roDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roDir, 0o755) }) // allow TempDir cleanup

	_, _, err := run(t, file)
	if err == nil {
		t.Fatal("expected a write error when the directory is not writable")
	}
}

// TestRunPreservesTrailingDataWithoutNewline checks the readFileToStrList branch
// where the final line has no trailing newline; it must still be converted and
// kept.
func TestRunPreservesTrailingDataWithoutNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "tail.txt")
	if err := os.WriteFile(file, []byte("a\nb"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, file); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	// The final "b" has no newline, so only the first LF becomes CRLF.
	if string(got) != "a\r\nb" {
		t.Errorf("file = %q, want %q", got, "a\r\nb")
	}
}

// TestEnvVarExpansionInOperand covers os.ExpandEnv on the operand.
func TestEnvVarExpansionInOperand(t *testing.T) {
	// No t.Parallel(): t.Setenv forbids it.
	dir := t.TempDir()
	file := filepath.Join(dir, "env.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("U2D_TESTDIR", dir)

	if _, _, err := run(t, "$U2D_TESTDIR/env.txt"); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, _ := os.ReadFile(file) //nolint:gosec // test-written file
	if string(got) != "x\r\n" {
		t.Errorf("file = %q, want CRLF after env expansion", got)
	}
}
