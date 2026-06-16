package grep_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAfterContext covers -A NUM: trailing context lines printed with the '-'
// separator.
func TestAfterContext(t *testing.T) {
	t.Parallel()
	in := "1\n2\nMATCH\n4\n5\n6\n"
	out, _, err := run(t, in, "-A2", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "MATCH\n4\n5\n"
	if out != want {
		t.Errorf("-A2 out = %q, want %q", out, want)
	}
}

// TestBeforeContext covers -B NUM: leading context lines printed before the
// match.
func TestBeforeContext(t *testing.T) {
	t.Parallel()
	in := "1\n2\n3\nMATCH\n5\n"
	out, _, err := run(t, in, "-B2", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "2\n3\nMATCH\n"
	if out != want {
		t.Errorf("-B2 out = %q, want %q", out, want)
	}
}

// TestContext covers -C NUM: surrounding context on both sides.
func TestContext(t *testing.T) {
	t.Parallel()
	in := "1\n2\n3\nMATCH\n5\n6\n7\n"
	out, _, err := run(t, in, "-C1", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "3\nMATCH\n5\n"
	if out != want {
		t.Errorf("-C1 out = %q, want %q", out, want)
	}
}

// TestContextWithLineNumbers verifies that context lines use '-' and match lines
// use ':' as the separator under -n.
func TestContextWithLineNumbers(t *testing.T) {
	t.Parallel()
	in := "a\nMATCH\nb\n"
	out, _, err := run(t, in, "-n", "-C1", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "1-a\n2:MATCH\n3-b\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestContextGroupSeparator verifies that two non-contiguous match groups are
// separated by a "--" line.
func TestContextGroupSeparator(t *testing.T) {
	t.Parallel()
	in := "MATCH\nb\nc\nd\ne\nf\nMATCH\n"
	out, _, err := run(t, in, "-A1", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "MATCH\nb\n--\nMATCH\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestContextOverlapNoSeparator verifies overlapping/contiguous groups merge
// without a "--" separator.
func TestContextOverlapNoSeparator(t *testing.T) {
	t.Parallel()
	in := "MATCH\nb\nMATCH\nd\n"
	out, _, err := run(t, in, "-A1", "MATCH")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// Trailing context of the first match (b) is followed immediately by the
	// second match, so no separator appears.
	want := "MATCH\nb\nMATCH\nd\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

// TestIncludeExclude covers --include and --exclude on a recursive search.
func TestIncludeExclude(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	txtFile := filepath.Join(dir, "notes.txt")
	logFile := filepath.Join(dir, "app.log")
	for _, f := range []string{goFile, txtFile, logFile} {
		if err := os.WriteFile(f, []byte("needle\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// --include only *.go
	out, _, err := run(t, "", "-r", "--include=*.go", "needle", dir)
	if err != nil {
		t.Fatalf("--include err = %v", err)
	}
	if !strings.Contains(out, goFile) || strings.Contains(out, txtFile) || strings.Contains(out, logFile) {
		t.Errorf("--include out = %q, want only %q", out, goFile)
	}

	// --exclude *.log
	out, _, err = run(t, "", "-r", "--exclude=*.log", "needle", dir)
	if err != nil {
		t.Fatalf("--exclude err = %v", err)
	}
	if strings.Contains(out, logFile) {
		t.Errorf("--exclude out = %q, should not contain %q", out, logFile)
	}
	if !strings.Contains(out, goFile) || !strings.Contains(out, txtFile) {
		t.Errorf("--exclude out = %q, want go and txt files", out)
	}
}

// TestExcludeDir covers --exclude-dir on a recursive search.
func TestExcludeDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	keep := filepath.Join(dir, "src")
	skip := filepath.Join(dir, "vendor")
	if err := os.MkdirAll(keep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skip, 0o755); err != nil {
		t.Fatal(err)
	}
	keepFile := filepath.Join(keep, "a.txt")
	skipFile := filepath.Join(skip, "b.txt")
	if err := os.WriteFile(keepFile, []byte("needle\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skipFile, []byte("needle\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "", "-r", "--exclude-dir=vendor", "needle", dir)
	if err != nil {
		t.Fatalf("--exclude-dir err = %v", err)
	}
	if !strings.Contains(out, keepFile) {
		t.Errorf("out = %q, want %q", out, keepFile)
	}
	if strings.Contains(out, skipFile) {
		t.Errorf("out = %q, should not contain %q (excluded dir)", out, skipFile)
	}
}

// TestColorAlways covers --color=always: the matched substring is wrapped in
// ANSI escapes.
func TestColorAlways(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello world\n", "--color=always", "world")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "hello \x1b[01;31mworld\x1b[0m\n"
	if out != want {
		t.Errorf("--color=always out = %q, want %q", out, want)
	}
}

// TestColorNever covers --color=never: no escapes emitted.
func TestColorNever(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "hello world\n", "--color=never", "world")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("--color=never out = %q, should have no escapes", out)
	}
}

// TestColorInvalid covers an invalid --color value exiting 2.
func TestColorInvalid(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "--color=bogus", "x")
	if code := exitCode(t, err); code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "grep:") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestByteOffset covers -b: the 0-based byte offset of each matching line.
func TestByteOffset(t *testing.T) {
	t.Parallel()
	// "aaa\n" is 4 bytes, so the second line begins at offset 4.
	out, _, err := run(t, "aaa\nbbb\nccc\n", "-b", "bbb")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "4:bbb\n"
	if out != want {
		t.Errorf("-b out = %q, want %q", out, want)
	}
}

// TestByteOffsetWithLineNumber verifies the prefix order: line number then byte
// offset.
func TestByteOffsetWithLineNumber(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "aaa\nbbb\n", "-n", "-b", "bbb")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "2:4:bbb\n"
	if out != want {
		t.Errorf("-nb out = %q, want %q", out, want)
	}
}

// TestFilesWithoutMatch covers -L: print only file names with NO match.
func TestFilesWithoutMatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	hit := filepath.Join(dir, "hit.txt")
	miss := filepath.Join(dir, "miss.txt")
	if err := os.WriteFile(hit, []byte("needle\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(miss, []byte("nothing here\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "", "-L", "needle", hit, miss)
	if err != nil {
		t.Fatalf("-L err = %v", err)
	}
	if strings.Contains(out, hit) {
		t.Errorf("-L out = %q, should not contain matching file %q", out, hit)
	}
	if !strings.Contains(out, miss) {
		t.Errorf("-L out = %q, want non-matching file %q", out, miss)
	}
}
