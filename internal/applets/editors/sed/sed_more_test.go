package sed_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReplacementTranslations drives translateRepl across its escape handling:
// the whole-match &, numbered group references, and the \n, \t, \\, \& and
// literal-$ escapes.
func TestReplacementTranslations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		script string
		in     string
		want   string
	}{
		{"whole match amp", `s/cat/[&]/`, "cat\n", "[cat]\n"},
		{"group ref", `s/\(a\)\(b\)/\2\1/`, "ab\n", "ba\n"},
		{"newline escape", `s/-/\n/`, "a-b\n", "a\nb\n"},
		{"tab escape", `s/-/\t/`, "a-b\n", "a\tb\n"},
		{"literal backslash", `s/x/\\/`, "x\n", "\\\n"},
		{"literal ampersand", `s/x/\&/`, "x\n", "&\n"},
		{"literal dollar in input", `s/a/Z/`, "a$\n", "Z$\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, errOut, err := run(t, tt.in, tt.script)
			if err != nil {
				t.Fatalf("err = %v (stderr=%q)", err, errOut)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestBREOperatorEscaping drives bre2ere: in basic regex mode the bare
// characters ( ) + ? { } | are literals, while their backslash-escaped forms are
// operators.
func TestBREOperatorEscaping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		script string
		in     string
		want   string
	}{
		// Bare '+' is a literal plus in BRE.
		{"literal plus", `s/a+/X/`, "a+b\n", "Xb\n"},
		// Escaped \+ is the one-or-more operator.
		{"operator plus", `s/a\+/X/`, "aaab\n", "Xb\n"},
		// Bare parentheses are literal in BRE.
		{"literal parens", `s/(x)/Y/`, "(x)\n", "Y\n"},
		// Escaped \( \) form a capture group whose match can be referenced in
		// the replacement.
		{"operator group", `s/\(ab\)c/[\1]/`, "abc\n", "[ab]\n"},
		// Bare '?' is literal.
		{"literal question", `s/a?/Q/`, "a?\n", "Q\n"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, errOut, err := run(t, tt.in, tt.script)
			if err != nil {
				t.Fatalf("err = %v (stderr=%q)", err, errOut)
			}
			if out != tt.want {
				t.Errorf("out = %q, want %q", out, tt.want)
			}
		})
	}
}

// TestExtendedRegexFlag drives compileRE/parse in extended mode, where bare
// operators keep their special meaning.
func TestExtendedRegexFlag(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "aaab\n", "-E", "s/a+/X/")
	if err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	if out != "Xb\n" {
		t.Errorf("out = %q, want %q", out, "Xb\n")
	}
}

// TestUnterminatedAddressRegex drives address()'s until-not-found path: a regex
// address with no closing delimiter is reported.
func TestUnterminatedAddressRegex(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "x\n", "/abc")
	if err == nil {
		t.Fatal("expected error for an unterminated address regex")
	}
	if !strings.Contains(errOut, "unterminated address regex") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestInPlaceMultipleFilesOneMissing drives runInPlace's per-file failure path:
// a missing file is reported but the readable file is still edited, and the run
// fails overall.
func TestInPlaceMultipleFilesOneMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	if err := os.WriteFile(good, []byte("foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.txt")

	_, errOut, err := run(t, "", "-i", "s/foo/bar/", missing, good)
	if err == nil {
		t.Fatal("expected an overall failure when one file is missing")
	}
	if !strings.Contains(errOut, "missing.txt") {
		t.Errorf("stderr = %q, want a diagnostic for the missing file", errOut)
	}
	got, rerr := os.ReadFile(good)
	if rerr != nil {
		t.Fatalf("read good: %v", rerr)
	}
	if string(got) != "bar\n" {
		t.Errorf("good file = %q, want %q", got, "bar\n")
	}
}

// TestInPlacePreservesMode drives writeFilePreservingMode: the in-place rewrite
// keeps the original permission bits of the file.
func TestInPlacePreservesMode(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningfully preserved on Windows")
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "perm.txt")
	if err := os.WriteFile(f, []byte("hello\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	// Ensure the mode is exactly 0640 even under a restrictive umask.
	if err := os.Chmod(f, 0o640); err != nil {
		t.Fatal(err)
	}

	if _, errOut, err := run(t, "", "-i", "s/hello/world/", f); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	info, err := os.Stat(f)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o640 {
		t.Errorf("mode after in-place edit = %o, want 640", perm)
	}
	got, _ := os.ReadFile(f)
	if string(got) != "world\n" {
		t.Errorf("content = %q, want %q", got, "world\n")
	}
}

// TestInPlaceRangeStateResetBetweenFiles drives cloneProgram: a two-address
// range that is open at the end of the first file must not leak into the second.
func TestInPlaceRangeStateResetBetweenFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	// In f1, the range "/start/,/end/d" never sees its end, so it stays open.
	if err := os.WriteFile(f1, []byte("keep\nstart\ndrop1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, errOut, err := run(t, "", "-i", "/start/,/end/d", f1, f2); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	g1, _ := os.ReadFile(f1)
	if string(g1) != "keep\n" {
		t.Errorf("f1 = %q, want %q", g1, "keep\n")
	}
	// f2 must be untouched: the range did not leak across files.
	g2, _ := os.ReadFile(f2)
	if string(g2) != "line1\nline2\n" {
		t.Errorf("f2 = %q, want it unchanged", g2)
	}
}
