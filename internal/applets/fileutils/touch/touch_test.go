package touch_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/touch"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := touch.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestRunCreatesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")

	_, errOut, err := run(t, path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("expected file to be created: %v", statErr)
	}
}

func TestRunUpdatesModTime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}

	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !info.ModTime().After(old) {
		t.Errorf("modtime = %v, want after %v", info.ModTime(), old)
	}
}

func TestRunNoCreate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.txt")

	_, errOut, err := run(t, "-c", path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("expected file to remain absent, stat err = %v", statErr)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if !strings.Contains(errOut, "touch: missing file operand") {
		t.Errorf("stderr = %q, want missing file operand message", errOut)
	}
}

func TestHelpSections(t *testing.T) {
	t.Parallel()
	out := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: &bytes.Buffer{}}
	if err := touch.New().Run(context.Background(), io, []string{"--help"}); err != nil {
		t.Fatalf("Run --help error = %v", err)
	}
	if !strings.Contains(out.String(), "Examples:") || !strings.Contains(out.String(), "Exit status:") {
		t.Errorf("--help missing structured sections:\n%s", out.String())
	}
}
