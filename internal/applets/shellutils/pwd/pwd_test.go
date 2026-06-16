package pwd_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/pwd"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := pwd.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunDefault(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	out, _, err := run(t)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("output should end with newline, got %q", out)
	}
	got := strings.TrimSuffix(out, "\n")

	// The output should name the current directory. On systems where the temp
	// directory contains symlinks (e.g. macOS /tmp -> /private/tmp), comparing
	// the canonical forms keeps the assertion robust.
	wantWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("output %q is not absolute", got)
	}
	if !sameDir(t, got, wantWD) {
		t.Errorf("output = %q, want current directory %q", got, wantWD)
	}
}

func TestRunPhysical(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	out, _, err := run(t, "-P")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	if !filepath.IsAbs(got) {
		t.Errorf("-P output %q is not absolute", got)
	}

	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != resolved {
		t.Errorf("-P output = %q, want resolved path %q", got, resolved)
	}
}

func TestRunLogicalHonoursPWD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PWD", dir)

	out, _, err := run(t, "-L")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got := strings.TrimSuffix(out, "\n")
	if !sameDir(t, got, dir) {
		t.Errorf("-L output = %q, want %q", got, dir)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("--help error = %v", err)
	}
	if !strings.Contains(out, "Usage: pwd") {
		t.Errorf("--help out = %q", out)
	}

	out, _, err = run(t, "--version")
	if err != nil {
		t.Fatalf("--version error = %v", err)
	}
	if !strings.Contains(out, "pwd (mimixbox)") {
		t.Errorf("--version out = %q", out)
	}
}

func sameDir(t *testing.T, a, b string) bool {
	t.Helper()
	ra, err := filepath.EvalSymlinks(a)
	if err != nil {
		return false
	}
	rb, err := filepath.EvalSymlinks(b)
	if err != nil {
		return false
	}
	return ra == rb
}

func TestHelpSections(t *testing.T) {
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := pwd.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Examples:") {
		t.Errorf("--help missing Examples section:\n%s", got)
	}
	if !strings.Contains(got, "Exit status:") {
		t.Errorf("--help missing Exit status section:\n%s", got)
	}
}
