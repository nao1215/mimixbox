package mkdir_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mkdir"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := mkdir.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := mkdir.New()
	if c.Name() != "mkdir" {
		t.Errorf("Name() = %q, want %q", c.Name(), "mkdir")
	}
	if c.Synopsis() != "Make directories" {
		t.Errorf("Synopsis() = %q, want %q", c.Synopsis(), "Make directories")
	}
}

func TestRunModeSetsPermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "secret")

	if _, errOut, err := run(t, "-m", "700", target); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	info, statErr := os.Stat(target)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("mode = %o, want 700", info.Mode().Perm())
	}
}

func TestRunInvalidMode(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "d")

	out, errOut, err := run(t, "-m", "notoctal", target)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "invalid mode") {
		t.Errorf("stderr = %q, want it to mention invalid mode", errOut)
	}
	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("directory must not be created on mode error")
	}
}

func TestRunVerbosePrintsCreated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "loud")

	out, errOut, err := run(t, "-v", target)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "created directory") || !strings.Contains(out, target) {
		t.Errorf("verbose out = %q, want it to mention the created directory", out)
	}
}

func TestRunCreatesDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "single")

	_, _, err := run(t, target)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, statErr := os.Stat(target)
	if statErr != nil {
		t.Fatalf("stat = %v", statErr)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", target)
	}
}

func TestRunParentsCreatesNested(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "parents", "child")

	_, _, err := run(t, "-p", target)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}
	info, statErr := os.Stat(target)
	if statErr != nil {
		t.Fatalf("stat = %v", statErr)
	}
	if !info.IsDir() {
		t.Errorf("%s is not a directory", target)
	}
}

func TestRunMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if strings.TrimRight(errOut, "\n") != "mkdir: no operand" {
		t.Errorf("stderr = %q, want %q", errOut, "mkdir: no operand")
	}
}

func TestRunExistingDirectoryWithoutParents(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "exists")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, target)
	if err == nil {
		t.Fatal("expected error for existing directory")
	}
	if errOut == "" {
		t.Errorf("expected error message on stderr")
	}
}

func TestRunExistingDirectoryWithParentsSucceeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "exists")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "-p", target)
	if err != nil {
		t.Fatalf("Run error = %v (want nil with -p on existing dir)", err)
	}
}
