package tee_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/tee"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := tee.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunStdoutAndFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")

	out, errOut, err := run(t, "hello\nworld\n", f)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if out != "hello\nworld\n" {
		t.Errorf("stdout = %q, want %q", out, "hello\nworld\n")
	}
	got, rerr := os.ReadFile(f)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(got) != "hello\nworld\n" {
		t.Errorf("file = %q, want %q", string(got), "hello\nworld\n")
	}
}

func TestRunMultipleFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")

	out, _, err := run(t, "data\n", a, b)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "data\n" {
		t.Errorf("stdout = %q, want %q", out, "data\n")
	}
	for _, f := range []string{a, b} {
		got, rerr := os.ReadFile(f)
		if rerr != nil {
			t.Fatal(rerr)
		}
		if string(got) != "data\n" {
			t.Errorf("file %s = %q, want %q", f, string(got), "data\n")
		}
	}
}

func TestRunAppend(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(f, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "second\n", "-a", f)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "second\n" {
		t.Errorf("stdout = %q, want %q", out, "second\n")
	}
	got, rerr := os.ReadFile(f)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(got) != "first\nsecond\n" {
		t.Errorf("file = %q, want %q", string(got), "first\nsecond\n")
	}
}

func TestRunTruncateByDefault(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	f := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(f, []byte("old content\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, "new\n", f); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, rerr := os.ReadFile(f)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(got) != "new\n" {
		t.Errorf("file = %q, want %q", string(got), "new\n")
	}
}

func TestRunInvalidFile(t *testing.T) {
	t.Parallel()
	// A path whose parent directory does not exist cannot be opened.
	bad := filepath.Join(t.TempDir(), "missing-dir", "out.txt")

	out, errOut, err := run(t, "payload\n", bad)
	if err == nil {
		t.Fatal("expected error for invalid file path")
	}
	if out != "payload\n" {
		t.Errorf("stdout = %q, want %q (stdout must still receive input)", out, "payload\n")
	}
	if !strings.Contains(errOut, "tee: "+bad+":") {
		t.Errorf("stderr = %q, want tee error prefix for %q", errOut, bad)
	}
}

func TestRunIgnoreInterruptsAccepted(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, "x\n", "-i")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if out != "x\n" {
		t.Errorf("stdout = %q, want %q", out, "x\n")
	}
}
