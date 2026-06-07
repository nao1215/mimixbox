package mktemp_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/debianutils/mktemp"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := mktemp.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNew(t *testing.T) {
	t.Parallel()
	if mktemp.New() == nil {
		t.Fatal("New() returned nil")
	}
}

func TestRunDefaultCreatesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, errOut, err := run(t, "-p", dir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	path := strings.TrimSpace(out)
	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Fatalf("path %q is not under %q", path, dir)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("expected file to exist: %v", statErr)
	}
	if !info.Mode().IsRegular() {
		t.Errorf("expected a regular file, got mode %v", info.Mode())
	}
	if rmErr := os.Remove(path); rmErr != nil {
		t.Errorf("cleanup failed: %v", rmErr)
	}
}

func TestRunDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, errOut, err := run(t, "-d", "-p", dir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	path := strings.TrimSpace(out)
	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Fatalf("path %q is not under %q", path, dir)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("expected directory to exist: %v", statErr)
	}
	if !info.IsDir() {
		t.Errorf("expected a directory, got mode %v", info.Mode())
	}
}

func TestRunDryRunDoesNotCreate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, errOut, err := run(t, "-u", "-p", dir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	path := strings.TrimSpace(out)
	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Fatalf("path %q is not under %q", path, dir)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("expected %q to NOT exist, stat err = %v", path, statErr)
	}
}

func TestRunTmpdirPlacement(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	out, errOut, err := run(t, "-p", dir, "myfile.XXXXXX")
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	path := strings.TrimSpace(out)
	if filepath.Dir(path) != dir {
		t.Fatalf("expected parent dir %q, got %q", dir, filepath.Dir(path))
	}
	if !strings.HasPrefix(filepath.Base(path), "myfile.") {
		t.Errorf("expected name to start with %q, got %q", "myfile.", filepath.Base(path))
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("expected file to exist: %v", statErr)
	}
	_ = os.Remove(path)
}

func TestRunTooFewX(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, "-p", dir, "bad.XX")
	if err == nil {
		t.Fatal("expected error for too few X's")
	}
	if !strings.Contains(errOut, "too few X's") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunTooManyTemplates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, "-p", dir, "a.XXXXXX", "b.XXXXXX")
	if err == nil {
		t.Fatal("expected error for too many templates")
	}
	if !strings.Contains(errOut, "too many templates") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunQuietSuppressesError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, errOut, err := run(t, "-q", "-p", dir, "bad.X")
	if err == nil {
		t.Fatal("expected error for too few X's")
	}
	if errOut != "" {
		t.Errorf("expected no stderr with -q, got %q", errOut)
	}
}
