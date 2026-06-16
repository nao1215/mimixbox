package which_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/debianutils/which"
)

func TestSynopsis(t *testing.T) {
	t.Parallel()
	if which.New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunAllExplicitPath drives lookPathAll's path-separator branch: a name that
// contains a separator is resolved directly (and made absolute) rather than
// searched on $PATH.
func TestRunAllExplicitPath(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "tool")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, "-a", bin)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got := strings.TrimSpace(out)
	if !filepath.IsAbs(got) {
		t.Errorf("out = %q, want an absolute path", got)
	}
	if got != bin {
		t.Errorf("out = %q, want %q", got, bin)
	}
}

// TestRunAllExplicitPathNotExecutable: an explicit path that is not an
// executable regular file yields nothing and a non-zero exit.
func TestRunAllExplicitPathNotExecutable(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(plain, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, code := execute(t, "-a", plain)
	if code == 0 {
		t.Error("a non-executable explicit path should make which exit non-zero")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

// TestRunAllNotFound covers the -a branch where a bare name matches nothing on
// $PATH.
func TestRunAllNotFound(t *testing.T) {
	out, _, code := execute(t, "-a", "this_command_does_not_exist_xyz")
	if code == 0 {
		t.Error("an unmatched -a name should make which exit non-zero")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}
