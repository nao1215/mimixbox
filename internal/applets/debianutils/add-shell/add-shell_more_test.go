package addShell_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	addShell "github.com/nao1215/mimixbox/internal/applets/debianutils/add-shell"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := addShell.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// TestNameSynopsis covers the Name and Synopsis metadata accessors.
func TestNameSynopsis(t *testing.T) {
	t.Parallel()
	c := addShell.New()
	if c.Name() != "add-shell" {
		t.Errorf("Name() = %q", c.Name())
	}
	if c.Synopsis() != "Add shell name to /etc/shells" {
		t.Errorf("Synopsis() = %q", c.Synopsis())
	}
}

// TestRunNoArgs covers the "no shell name given" branch of Run.
func TestRunNoArgs(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t)
	if err == nil {
		t.Fatal("expected error when no shell name is given")
	}
	if !strings.Contains(errOut, "add-shell:") {
		t.Errorf("stderr = %q, want add-shell usage", errOut)
	}
}

// TestRunWriteFailure drives Run's addShells error branch. Run always targets
// /etc/shells, which an unprivileged test cannot write; the operation must fail
// and Run must report it and exit non-zero. If the test happens to run as root
// (where the append could succeed) it is skipped so it stays deterministic.
func TestRunWriteFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; /etc/shells may be writable")
	}
	// A shell that is almost certainly not already listed, so addShells reaches
	// the append (and the open fails on a read-only /etc/shells).
	_, errOut, err := run(t, "/nonexistent/mimixbox-test-shell")
	if err == nil {
		t.Skip("environment allowed writing /etc/shells; nothing to assert")
	}
	if !strings.Contains(errOut, "add-shell:") {
		t.Errorf("stderr = %q, want add-shell error prefix", errOut)
	}
}

// TestAddShellsCreatesMissingFile covers readShells' os.IsNotExist branch (a
// not-yet-created shells file is treated as empty) and the file-creation path
// of addShells.
func TestAddShellsCreatesMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells") // does not exist yet

	if err := addShell.AddShellsForTest(path, []string{"/bin/sh", "/bin/zsh"}); err != nil {
		t.Fatalf("addShells error = %v", err)
	}
	got := readLines(t, path)
	if strings.Join(got, ",") != "/bin/sh,/bin/zsh" {
		t.Errorf("lines = %v, want [/bin/sh /bin/zsh]", got)
	}
}

// TestAddShellsSkipsBlankAndDuplicate covers readShells' blank-line skipping and
// the in-batch duplicate suppression in addShells.
func TestAddShellsSkipsBlankAndDuplicate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte("/bin/sh\n\n  \n/bin/bash\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// /bin/bash already present; /bin/dash given twice should be added once.
	if err := addShell.AddShellsForTest(path, []string{"/bin/bash", "/bin/dash", "/bin/dash"}); err != nil {
		t.Fatalf("addShells error = %v", err)
	}
	// The pre-existing blank lines are left untouched; /bin/dash is appended
	// exactly once and /bin/bash (already listed) is not re-added.
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(content), "/bin/dash"); n != 1 {
		t.Errorf("/bin/dash appears %d times, want 1: %q", n, content)
	}
	if n := strings.Count(string(content), "/bin/bash"); n != 1 {
		t.Errorf("/bin/bash appears %d times, want 1 (no duplicate): %q", n, content)
	}
}

// TestAddShellsReadError covers addShells returning readShells' error when the
// shells path is a directory rather than a regular file.
func TestAddShellsReadError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir() // a directory used where a file is expected
	if err := addShell.AddShellsForTest(dir, []string{"/bin/sh"}); err == nil {
		t.Error("expected error when shells path is a directory")
	}
}
