package base32_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEncodeNoWrap covers wrapLines' cols <= 0 branch (-w 0 disables wrapping):
// a long encoding must come back as a single line plus a trailing newline.
func TestEncodeNoWrap(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, strings.Repeat("a", 100), "-w", "0")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	body := strings.TrimRight(out, "\n")
	if strings.Contains(body, "\n") {
		t.Errorf("-w 0 should not wrap, got %d lines: %q", strings.Count(body, "\n")+1, out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output should end with a newline: %q", out)
	}
}

// TestEncodeFile covers operand() returning a named FILE (not "-") and encoding
// data read from disk.
func TestEncodeFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "in.bin")
	if err := os.WriteFile(f, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := run(t, "", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if strings.TrimRight(out, "\n") != "NBSWY3DP" {
		t.Errorf("encode of file = %q, want NBSWY3DP", out)
	}
}

// TestDecodeWhitespaceStripped covers the default (non -i) decode path where
// embedded ASCII whitespace is stripped before decoding.
func TestDecodeWhitespaceStripped(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "NBSWY3DP\n", "-d")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "hello" {
		t.Errorf("decode = %q, want hello", out)
	}
}
