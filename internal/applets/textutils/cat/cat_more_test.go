package cat_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/textutils/cat"
)

// TestSynopsis covers the Synopsis metadata accessor, which the Run-driven
// tests never call.
func TestSynopsis(t *testing.T) {
	if got := cat.New().Synopsis(); got != "Concatenate files and print on the standard output" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestRunMultipleMissingFiles drives keep()'s "existing != nil" branch: two
// failed opens must still yield a single failure and report both names, while
// the good file in between is still printed.
func TestRunMultipleMissingFiles(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	miss1 := filepath.Join(dir, "miss1.txt")
	miss2 := filepath.Join(dir, "miss2.txt")

	out, errOut, err := runStdin(t, "", miss1, good, miss2)
	if err == nil {
		t.Fatal("expected error when files are missing")
	}
	if out != "ok\n" {
		t.Errorf("out = %q, want the readable file's contents", out)
	}
	if !strings.Contains(errOut, "miss1.txt") || !strings.Contains(errOut, "miss2.txt") {
		t.Errorf("stderr = %q, want both missing files reported", errOut)
	}
}

// TestRunSqueezeShowEndsTabsCombined exercises renderStream with all of -s, -E
// and -T together, including the squeeze that drops a repeated blank line.
func TestRunSqueezeShowEndsTabsCombined(t *testing.T) {
	out, _, err := runStdin(t, "a\tb\n\n\n\nc\n", "-s", "-E", "-T")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	want := "a^Ib$\n$\nc$\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestRunNoTrailingNewline checks renderStream's handling of a final chunk that
// lacks a trailing newline: -E marks where a line ends, and GNU cat writes no
// $ where there is no newline, numbered or not.
func TestRunNoTrailingNewline(t *testing.T) {
	t.Parallel()
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"-E"}, "a$\ntail"},
		{[]string{"-A"}, "a$\ntail"},
		{[]string{"-n", "-E"}, "     1\ta$\n     2\ttail"},
	}
	for _, tt := range tests {
		out, _, err := runStdin(t, "a\ntail", tt.args...)
		if err != nil {
			t.Fatalf("Run error = %v", err)
		}
		if out != tt.want {
			t.Errorf("cat %v = %q, want %q", tt.args, out, tt.want)
		}
	}
}
