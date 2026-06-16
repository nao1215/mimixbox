package unexpand_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/unexpand"
)

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if unexpand.New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestLiteralTabRealignsBlanks drives the '\t' case of convertLine, where an
// existing tab in the input advances the column so following spaces realign to
// the next tab stop.
func TestLiteralTabRealignsBlanks(t *testing.T) {
	t.Parallel()
	// A literal tab at column 0 advances to column 8; the following 8 spaces then
	// span one more tab stop and collapse to a single tab.
	in := "\t        x\n"
	out, _, err := runStdin(t, in, "-a")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "\t\tx\n" {
		t.Errorf("out = %q, want %q", out, "\t\tx\n")
	}
}

// TestPartialTabStopStaysSpaces verifies that trailing blanks that do not reach
// the next tab stop are emitted as spaces, not tabs (the "remaining columns"
// loop of convertLine).
func TestPartialTabStopStaysSpaces(t *testing.T) {
	t.Parallel()
	// 8 leading spaces become one tab; the remaining 3 spaces do not reach the
	// next stop and stay as spaces.
	out, _, err := runStdin(t, "           y\n")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "\t   y\n" {
		t.Errorf("out = %q, want %q", out, "\t   y\n")
	}
}

// TestFileOperandConverts exercises the file (non-stdin) path of run.
func TestFileOperandConverts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(f, []byte("        x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := runStdin(t, "", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "\tx\n" {
		t.Errorf("out = %q, want %q", out, "\tx\n")
	}
}

// TestTwoFilesOneMissingKeepsFirstError exercises the keep() helper: the first
// error is preserved while later files still process.
func TestTwoFilesOneMissingKeepsFirstError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("        z\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, errOut, err := runStdin(t, "", "/no/such/file", good)
	if err == nil {
		t.Fatal("expected a non-zero exit because one file is missing")
	}
	if !strings.Contains(errOut, "unexpand: /no/such/file:") {
		t.Errorf("stderr = %q, want the missing-file message", errOut)
	}
	// The good file is still converted to stdout.
	if out != "\tz\n" {
		t.Errorf("out = %q, want the good file converted", out)
	}
}
