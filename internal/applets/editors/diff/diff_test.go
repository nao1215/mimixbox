package diff_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/editors/diff"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := diff.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var ee *command.ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return -1
}

// files writes two temp files with the given contents and returns their paths.
func files(t *testing.T, a, b string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	pa := filepath.Join(dir, "a")
	pb := filepath.Join(dir, "b")
	if err := os.WriteFile(pa, []byte(a), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pb, []byte(b), 0o644); err != nil {
		t.Fatal(err)
	}
	return pa, pb
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := diff.New()
	if got := c.Name(); got != "diff" {
		t.Errorf("Name() = %q", got)
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis empty")
	}
}

func TestIdenticalExit0(t *testing.T) {
	t.Parallel()
	a, b := files(t, "one\ntwo\n", "one\ntwo\n")
	out, _, err := run(t, a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if exitCode(t, err) != 0 {
		t.Errorf("exit = %d, want 0", exitCode(t, err))
	}
}

func TestChangeNormal(t *testing.T) {
	t.Parallel()
	a, b := files(t, "one\ntwo\nthree\n", "one\n2\nthree\n")
	out, _, err := run(t, a, b)
	if exitCode(t, err) != 1 {
		t.Fatalf("exit = %d, want 1", exitCode(t, err))
	}
	want := "2c2\n< two\n---\n> 2\n"
	if out != want {
		t.Errorf("normal diff:\n got %q\nwant %q", out, want)
	}
}

func TestDeletionNormal(t *testing.T) {
	t.Parallel()
	a, b := files(t, "one\ntwo\nthree\n", "one\nthree\n")
	out, _, err := run(t, a, b)
	if exitCode(t, err) != 1 {
		t.Fatalf("exit = %d", exitCode(t, err))
	}
	want := "2d1\n< two\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestInsertionNormal(t *testing.T) {
	t.Parallel()
	a, b := files(t, "one\nthree\n", "one\ntwo\nthree\n")
	out, _, err := run(t, a, b)
	if exitCode(t, err) != 1 {
		t.Fatalf("exit = %d", exitCode(t, err))
	}
	want := "1a2\n> two\n"
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestUnified(t *testing.T) {
	t.Parallel()
	a, b := files(t, "one\ntwo\nthree\n", "one\n2\nthree\n")
	out, _, err := run(t, "-u", a, b)
	if exitCode(t, err) != 1 {
		t.Fatalf("exit = %d", exitCode(t, err))
	}
	if !strings.HasPrefix(out, "--- ") {
		t.Errorf("missing --- header: %q", out)
	}
	if !strings.Contains(out, "@@ -") {
		t.Errorf("missing @@ hunk: %q", out)
	}
	if !strings.Contains(out, "-two") || !strings.Contains(out, "+2") {
		t.Errorf("unified body wrong: %q", out)
	}
	if !strings.Contains(out, " one") || !strings.Contains(out, " three") {
		t.Errorf("unified context wrong: %q", out)
	}
}

func TestBrief(t *testing.T) {
	t.Parallel()
	a, b := files(t, "x\n", "y\n")
	out, _, err := run(t, "-q", a, b)
	if exitCode(t, err) != 1 {
		t.Fatalf("exit = %d", exitCode(t, err))
	}
	if !strings.Contains(out, "differ") {
		t.Errorf("brief out = %q", out)
	}
}

func TestBriefIdentical(t *testing.T) {
	t.Parallel()
	a, b := files(t, "same\n", "same\n")
	out, _, err := run(t, "-q", a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
}

func TestIgnoreCase(t *testing.T) {
	t.Parallel()
	a, b := files(t, "Hello\n", "hello\n")
	out, _, err := run(t, "-i", a, b)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if out != "" {
		t.Errorf("with -i, out = %q, want empty", out)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "only-one")
	if exitCode(t, err) != 2 {
		t.Errorf("exit = %d, want 2", exitCode(t, err))
	}
	if !strings.Contains(errOut, "missing operand") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestExtraOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, "a", "b", "c")
	if exitCode(t, err) != 2 {
		t.Errorf("exit = %d, want 2", exitCode(t, err))
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want 'extra operand'", errOut)
	}
}

func TestMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a")
	if err := os.WriteFile(a, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, a, filepath.Join(dir, "nope"))
	if exitCode(t, err) != 2 {
		t.Errorf("exit = %d, want 2", exitCode(t, err))
	}
	if !strings.Contains(errOut, "diff:") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "--help")
	if err != nil {
		t.Fatalf("help err = %v", err)
	}
	if !strings.Contains(out, "Usage: diff") {
		t.Errorf("help = %q", out)
	}
	for _, want := range []string{"Examples:", "Exit status:"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q:\n%s", want, out)
		}
	}
}
