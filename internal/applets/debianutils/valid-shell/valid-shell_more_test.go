package validShell_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	validShell "github.com/nao1215/mimixbox/internal/applets/debianutils/valid-shell"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(""), Out: out, Err: errBuf}
	err := validShell.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

// TestSynopsis covers the Synopsis accessor.
func TestSynopsis(t *testing.T) {
	t.Parallel()
	if validShell.New().Synopsis() == "" {
		t.Error("Synopsis() is empty")
	}
}

// TestRunWithValidFileSucceeds drives Run end-to-end against a FILE operand that
// lists an existing, executable shell. This covers the success exit path of Run
// that the helper-only tests leave at 0%.
func TestRunWithValidFileSucceeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	// /bin/sh exists and is executable on every system we run on. Comment and
	// blank lines are also present to exercise those skip branches.
	if err := os.WriteFile(path, []byte("# a comment\n\n/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, errOut, err := run(t, path)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "OK: /bin/sh") {
		t.Errorf("stdout = %q, want an OK line", out)
	}
	if strings.Contains(out, "# a comment") {
		t.Errorf("comment lines must be ignored, got %q", out)
	}
}

// TestRunWithInvalidShellExitsNonZero covers the !ok branch of Run, where a
// listed shell does not exist.
func TestRunWithInvalidShellExitsNonZero(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte("/no/such/shell\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, path)
	if err == nil {
		t.Fatal("expected a non-zero exit for a missing shell")
	}
	if !strings.Contains(out, "NG: /no/such/shell") {
		t.Errorf("stdout = %q, want an NG line", out)
	}
}

// TestRunWithMissingFileReportsError covers the open-error branch of Run.
func TestRunWithMissingFileReportsError(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, filepath.Join(t.TempDir(), "no-shells-file"))
	if err == nil {
		t.Fatal("expected an error for a missing shells file")
	}
	if !strings.Contains(errOut, "valid-shell:") {
		t.Errorf("stderr = %q, want a valid-shell error prefix", errOut)
	}
}

// TestValidateShellsSkipsCommentsAndBlanks covers the comment/blank skip
// branches of validateShells via the exported test hook.
func TestValidateShellsSkipsCommentsAndBlanks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte("# header\n\n   \n/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ok, err := validShell.ValidateShellsForTest(path, &out)
	if err != nil {
		t.Fatalf("validateShells error = %v", err)
	}
	if !ok {
		t.Errorf("ok = false, want true; output = %q", out.String())
	}
	// Only the /bin/sh line produces output.
	if lines := strings.Count(strings.TrimSpace(out.String()), "\n"); lines != 0 {
		t.Errorf("expected a single output line, got:\n%s", out.String())
	}
}

// TestValidateShellsDirectoryEntryIsNotExecutable covers isExecutable's IsDir
// branch: a directory listed as a shell is reported NG.
func TestValidateShellsDirectoryEntryIsNotExecutable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	subdir := filepath.Join(dir, "adir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte(subdir+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ok, err := validShell.ValidateShellsForTest(path, &out)
	if err != nil {
		t.Fatalf("validateShells error = %v", err)
	}
	if ok {
		t.Error("a directory must not count as an executable shell")
	}
	if !strings.Contains(out.String(), "NG: "+subdir) {
		t.Errorf("output = %q, want an NG line for the directory", out.String())
	}
}

// TestValidateShellsNonExecutableFileIsNG covers isExecutable's permission-bit
// branch: an existing regular file without an execute bit is NG.
func TestValidateShellsNonExecutableFileIsNG(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	plain := filepath.Join(dir, "plain")
	if err := os.WriteFile(plain, []byte("not a shell"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "shells")
	if err := os.WriteFile(path, []byte(plain+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	ok, err := validShell.ValidateShellsForTest(path, &out)
	if err != nil {
		t.Fatalf("validateShells error = %v", err)
	}
	if ok {
		t.Error("a non-executable file must not count as a valid shell")
	}
	if !strings.Contains(out.String(), "NG: "+plain) {
		t.Errorf("output = %q, want an NG line", out.String())
	}
}
