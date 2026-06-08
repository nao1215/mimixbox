package shred_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/shred"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := shred.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func tmpFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestOverwriteKeepsSize(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "0123456789")
	if _, _, err := run(t, "-n", "2", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 10 {
		t.Errorf("size = %d, want 10 (file should still exist)", info.Size())
	}
}

func TestZeroPass(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "abcdef")
	if _, _, err := run(t, "-n", "0", "-z", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	got, err := os.ReadFile(p) //nolint:gosec // reading a file the test created
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, make([]byte, 6)) {
		t.Errorf("after zero pass = %v, want all zeros", got)
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "data")
	if _, _, err := run(t, "-u", p); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("file should have been removed")
	}
}

func TestVerbose(t *testing.T) {
	t.Parallel()
	p := tmpFile(t, "x")
	_, errOut, err := run(t, "-n", "1", "-v", p)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if !strings.Contains(errOut, "pass 1/1") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestDirectoryRejected(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errOut, "is a directory") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestInvalidIterations(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, "-n", "-1", "file")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid number of passes") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing file operand") {
		t.Errorf("err = %v", err)
	}
}

func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := shred.New()
	if c.Name() != "shred" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}
