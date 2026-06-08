package builtin

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/command"
)

func newIO(in string) (command.IO, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	return command.IO{In: strings.NewReader(in), Out: out, Err: errBuf}, out, errBuf
}

func TestIsBuiltinCmd(t *testing.T) {
	t.Parallel()
	if !IsBuiltinCmd("cd") {
		t.Error("cd should be a builtin")
	}
	if IsBuiltinCmd("ls") {
		t.Error("ls should not be a builtin")
	}
	if IsBuiltinCmd("") {
		t.Error("empty string should not be a builtin")
	}
}

func TestRunNotBuiltin(t *testing.T) {
	t.Parallel()
	io, _, _ := newIO("")
	if err := Run(io, "nope", nil); !errors.Is(err, ErrNotBuiltinCmd) {
		t.Errorf("Run(nope) err = %v, want ErrNotBuiltinCmd", err)
	}
}

func TestCdToDirectory(t *testing.T) {
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })

	dir := t.TempDir()
	want, _ := filepath.EvalSymlinks(dir)
	io, _, _ := newIO("")
	if err := Run(io, "cd", []string{dir}); err != nil {
		t.Fatalf("cd err = %v", err)
	}
	got, _ := os.Getwd()
	got, _ = filepath.EvalSymlinks(got)
	if got != want {
		t.Errorf("cwd = %q, want %q", got, want)
	}
	// PWD should track the new directory.
	if pwd, _ := filepath.EvalSymlinks(os.Getenv("PWD")); pwd != want {
		t.Errorf("PWD = %q, want %q", pwd, want)
	}
}

func TestCdNoArgUsesHome(t *testing.T) {
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })

	home := t.TempDir()
	t.Setenv("HOME", home)
	want, _ := filepath.EvalSymlinks(home)

	io, _, _ := newIO("")
	if err := Run(io, "cd", nil); err != nil {
		t.Fatalf("cd err = %v", err)
	}
	got, _ := os.Getwd()
	got, _ = filepath.EvalSymlinks(got)
	if got != want {
		t.Errorf("cwd = %q, want HOME %q", got, want)
	}
}

func TestCdNoHome(t *testing.T) {
	t.Setenv("HOME", "")
	io, _, _ := newIO("")
	if err := Run(io, "cd", nil); !errors.Is(err, ErrNoHome) {
		t.Errorf("err = %v, want ErrNoHome", err)
	}
}

func TestCdDash(t *testing.T) {
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })

	start, _ := filepath.EvalSymlinks(orig)
	dir := t.TempDir()

	io, out, _ := newIO("")
	if err := Run(io, "cd", []string{dir}); err != nil {
		t.Fatal(err)
	}
	if err := Run(io, "cd", []string{"-"}); err != nil {
		t.Fatalf("cd - err = %v", err)
	}
	got, _ := os.Getwd()
	got, _ = filepath.EvalSymlinks(got)
	if got != start {
		t.Errorf("cd - cwd = %q, want %q", got, start)
	}
	// cd - echoes the directory it moved to.
	if strings.TrimSpace(out.String()) == "" {
		t.Error("cd - should print the directory")
	}
}

func TestCdDashNoOldpwd(t *testing.T) {
	t.Setenv("OLDPWD", "")
	io, _, _ := newIO("")
	if err := Run(io, "cd", []string{"-"}); !errors.Is(err, ErrNoOldpwd) {
		t.Errorf("err = %v, want ErrNoOldpwd", err)
	}
}

func TestCdNonexistent(t *testing.T) {
	io, _, _ := newIO("")
	err := Run(io, "cd", []string{filepath.Join(t.TempDir(), "does-not-exist")})
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
