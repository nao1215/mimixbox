package patch_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/editors/patch"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := patch.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := patch.New()
	if got := c.Name(); got != "patch" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

// unifiedPatch returns a unified diff that turns "one\ntwo\nthree\n" into
// "one\n2\nthree\n" for the file named target.
func unifiedPatch(target string) string {
	return "--- " + target + "\n" +
		"+++ " + target + "\n" +
		"@@ -1,3 +1,3 @@\n" +
		" one\n" +
		"-two\n" +
		"+2\n" +
		" three\n"
}

func TestApplyUnified(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, errOut, err := run(t, unifiedPatch(target)); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "one\n2\nthree\n" {
		t.Errorf("patched = %q", got)
	}
}

func TestApplyFromInputFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	patchFile := filepath.Join(dir, "p.diff")
	if err := os.WriteFile(patchFile, []byte(unifiedPatch(target)), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, "", "-i", patchFile); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "one\n2\nthree\n" {
		t.Errorf("patched = %q", got)
	}
}

func TestReverse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	// Already-patched file; -R should undo it back to the original.
	if err := os.WriteFile(target, []byte("one\n2\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, unifiedPatch(target), "-R"); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "one\ntwo\nthree\n" {
		t.Errorf("reversed = %q", got)
	}
}

func TestStrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("one\ntwo\nthree\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Patch names the file as a/<target> and b/<target>; -p1 strips the prefix.
	diff := "--- a/" + target + "\n" +
		"+++ b/" + target + "\n" +
		"@@ -1,3 +1,3 @@\n one\n-two\n+2\n three\n"
	if _, errOut, err := run(t, diff, "-p1"); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "one\n2\nthree\n" {
		t.Errorf("patched = %q", got)
	}
}

func TestDryRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	orig := "one\ntwo\nthree\n"
	if err := os.WriteFile(target, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, unifiedPatch(target), "--dry-run"); err != nil {
		t.Fatalf("err = %v", err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != orig {
		t.Errorf("--dry-run modified the file: %q", got)
	}
}

func TestAddAndDelete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Delete b, add X after c.
	diff := "--- " + target + "\n+++ " + target + "\n" +
		"@@ -1,3 +1,3 @@\n a\n-b\n c\n+X\n"
	if _, errOut, err := run(t, diff); err != nil {
		t.Fatalf("err = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "a\nc\nX\n" {
		t.Errorf("patched = %q, want a\\nc\\nX", got)
	}
}

func TestContextMismatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(target, []byte("totally\ndifferent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, unifiedPatch(target))
	if err == nil {
		t.Error("expected error on context mismatch")
	}
	if !strings.Contains(errOut, "patch:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	t.Run("empty input", func(t *testing.T) {
		t.Parallel()
		_, errOut, err := run(t, "")
		if err == nil {
			t.Error("expected error for empty patch")
		}
		if !strings.Contains(errOut, "no valid patches") {
			t.Errorf("stderr = %q", errOut)
		}
	})
	t.Run("missing input file", func(t *testing.T) {
		t.Parallel()
		_, errOut, err := run(t, "", "-i", "/no/such/patch/file")
		if err == nil {
			t.Error("expected error for missing -i file")
		}
		if !strings.Contains(errOut, "patch:") {
			t.Errorf("stderr = %q", errOut)
		}
	})
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "", "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: patch") {
		t.Errorf("help = %q", out)
	}
}
