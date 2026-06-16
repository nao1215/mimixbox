package wc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/wc"
)

func TestSynopsis(t *testing.T) {
	t.Parallel()
	c := wc.New()
	if c.Name() != "wc" {
		t.Errorf("Name() = %q, want wc", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunMaxLineLength drives the -L column (the maxLine branch of
// selectedValues and formatRow), checking the longest-line width is reported.
func TestRunMaxLineLength(t *testing.T) {
	t.Parallel()
	// Longest line is "hello" -> width 5; trailing newline ignored.
	out, _, err := run(t, "ab\nhello\nc\n", "-L")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "5\n" {
		t.Errorf("out = %q, want %q", out, "5\n")
	}
}

// TestRunAllColumns exercises every selected column together so the multi-column
// width logic (fieldWidth with unknownSize from stdin) is covered.
func TestRunAllColumns(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "alpha beta\n", "-l", "-w", "-m", "-c", "-L")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// lines=1 words=2 chars=11 bytes=11 maxline=10, padded to width 7.
	fields := strings.Fields(out)
	if len(fields) != 5 {
		t.Fatalf("out = %q, want 5 columns", out)
	}
	want := []string{"1", "2", "11", "11", "10"}
	for i, w := range want {
		if fields[i] != w {
			t.Errorf("column %d = %q, want %q (out=%q)", i, fields[i], w, out)
		}
	}
}

// TestRunMultipleMissingFiles makes two files fail so the keep() error-chaining
// branch (an already-set firstErr is preserved) is exercised, and a total line
// is not emitted because no row succeeded.
func TestRunMultipleMissingFiles(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "/no/such/a", "/no/such/b")
	if err == nil {
		t.Fatal("expected error for missing files")
	}
	if !strings.Contains(errOut, "/no/such/a") || !strings.Contains(errOut, "/no/such/b") {
		t.Errorf("stderr = %q, want both missing files reported", errOut)
	}
	if out != "" {
		t.Errorf("out = %q, want empty when all inputs failed", out)
	}
}

// TestRunDirectory covers the path where a directory opens but cannot be read:
// wc prints a zero row, reports the error, and exits non-zero.
func TestRunDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(f, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, "", "-l", dir, f)
	if err == nil {
		t.Fatal("expected error for an unreadable directory input")
	}
	if !strings.Contains(errOut, "wc:") {
		t.Errorf("stderr = %q, want wc: prefix", errOut)
	}
	// The good file's row and a total line should still be printed, widened to 7
	// because the directory marks the size as unknown.
	if !strings.Contains(out, "total") {
		t.Errorf("out = %q, want a total line", out)
	}
	if !strings.Contains(out, f) {
		t.Errorf("out = %q, want the readable file row", out)
	}
}
