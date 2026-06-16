package which_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/debianutils/which"
	"github.com/nao1215/mimixbox/internal/command"
)

// run executes the which command directly and returns stdout, stderr and the
// returned error.
func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := which.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// execute runs the which command through command.Execute and returns stdout,
// stderr and the resulting process exit code.
func execute(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	code := command.Execute(context.Background(), which.New(), io, args)
	return out.String(), errBuf.String(), code
}

// setupExecutable creates an executable named name in a fresh temp directory and
// prepends that directory to PATH for the duration of the test.
func setupExecutable(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return path
}

func TestNew(t *testing.T) {
	t.Parallel()
	c := which.New()
	if c == nil {
		t.Fatal("New returned nil")
	}
	if c.Name() != "which" {
		t.Errorf("Name = %q, want %q", c.Name(), "which")
	}
}

func TestRunFound(t *testing.T) {
	want := setupExecutable(t, "mimixbox-test-bin")

	got, errOut, err := run(t, "mimixbox-test-bin")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	lookup, lerr := exec.LookPath("mimixbox-test-bin")
	if lerr != nil {
		t.Fatalf("exec.LookPath: %v", lerr)
	}
	if got != lookup+"\n" {
		t.Errorf("out = %q, want %q (exec.LookPath)", got, lookup+"\n")
	}
	if got != want+"\n" {
		t.Errorf("out = %q, want %q", got, want+"\n")
	}
}

func TestRunNotFound(t *testing.T) {
	out, errOut, code := execute(t, "this_command_does_not_exist_xyz")
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
}

func TestRunMixed(t *testing.T) {
	bin := setupExecutable(t, "mimixbox-mixed-bin")

	out, _, code := execute(t, "mimixbox-mixed-bin", "this_command_does_not_exist_xyz")
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if out != bin+"\n" {
		t.Errorf("out = %q, want %q", out, bin+"\n")
	}
}

func TestRunNoOperand(t *testing.T) {
	out, errOut, code := execute(t)
	if code != command.ExitFailure {
		t.Errorf("exit code = %d, want %d", code, command.ExitFailure)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty", errOut)
	}
}

func TestRunAll(t *testing.T) {
	bin := setupExecutable(t, "mimixbox-all-bin")

	out, _, err := run(t, "-a", "mimixbox-all-bin")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(out, bin) {
		t.Errorf("out = %q, want to contain %q", out, bin)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := which.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("help err = %v", err)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("--help output missing %q:\n%s", want, out.String())
		}
	}
}
