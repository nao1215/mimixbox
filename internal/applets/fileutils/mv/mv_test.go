package mv_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/fileutils/mv"
	"github.com/nao1215/mimixbox/internal/command"
)

func run(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io := command.IO{In: strings.NewReader(stdin), Out: out, Err: errBuf}
	err := mv.New().Run(context.Background(), io, args)
	return out.String(), errBuf.String(), err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestNameAndSynopsis(t *testing.T) {
	t.Parallel()
	c := mv.New()
	if c.Name() != "mv" {
		t.Errorf("Name() = %q, want %q", c.Name(), "mv")
	}
	want := "Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY"
	if c.Synopsis() != want {
		t.Errorf("Synopsis() = %q, want %q", c.Synopsis(), want)
	}
}

func TestRename(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	writeFile(t, src, "hello\n")

	_, errOut, err := run(t, "", src, dest)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if _, statErr := os.Stat(src); !os.IsNotExist(statErr) {
		t.Errorf("source %s should no longer exist", src)
	}
	got, readErr := os.ReadFile(dest)
	if readErr != nil {
		t.Fatalf("reading dest: %v", readErr)
	}
	if string(got) != "hello\n" {
		t.Errorf("dest content = %q, want %q", string(got), "hello\n")
	}
}

func TestMoveIntoDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, src, "data\n")

	_, errOut, err := run(t, "", src, destDir)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	moved := filepath.Join(destDir, "file.txt")
	got, readErr := os.ReadFile(moved)
	if readErr != nil {
		t.Fatalf("reading moved file: %v", readErr)
	}
	if string(got) != "data\n" {
		t.Errorf("moved content = %q, want %q", string(got), "data\n")
	}
	if _, statErr := os.Stat(src); !os.IsNotExist(statErr) {
		t.Errorf("source %s should no longer exist", src)
	}
}

func TestNoClobberDoesNotOverwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "file.txt")
	destDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(destDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, src, "new\n")
	existing := filepath.Join(destDir, "file.txt")
	writeFile(t, existing, "old\n")

	if _, _, err := run(t, "", "-n", src, destDir); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	// -n must keep the existing destination untouched.
	got, readErr := os.ReadFile(existing)
	if readErr != nil {
		t.Fatalf("reading existing file: %v", readErr)
	}
	if string(got) != "old\n" {
		t.Errorf("dest content = %q, want %q (must not overwrite)", string(got), "old\n")
	}
}

func TestVerboseReportsRename(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dest := filepath.Join(dir, "b.txt")
	writeFile(t, src, "x\n")

	out, errOut, err := run(t, "", "-v", src, dest)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	if !strings.Contains(out, "renamed") || !strings.Contains(out, dest) {
		t.Errorf("verbose output = %q, want it to mention the rename to %q", out, dest)
	}
}

func TestForceOverwritesViaRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dest := filepath.Join(dir, "b.txt")
	writeFile(t, src, "new\n")
	writeFile(t, dest, "old\n")

	if _, errOut, err := run(t, "", "-f", src, dest); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(dest) //nolint:gosec // test-written file
	if string(got) != "new\n" {
		t.Errorf("dest content = %q, want overwrite to %q", got, "new\n")
	}
}

func TestBackupCreatesBackupViaRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dest := filepath.Join(dir, "b.txt")
	writeFile(t, src, "new\n")
	writeFile(t, dest, "old\n")

	// -b -i with a "yes" answer triggers the backup+interactive force path,
	// where the source is moved aside to the "~" suffixed name rather than
	// over the existing destination, so the original destination survives.
	if _, errOut, err := run(t, "y\n", "-b", "-i", src, dest); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	backup := dest + "~"
	movedContent, err := os.ReadFile(backup) //nolint:gosec // test-written file
	if err != nil {
		t.Fatalf("expected backup-named file %s: %v", backup, err)
	}
	if string(movedContent) != "new\n" {
		t.Errorf("backup-named file content = %q, want %q", movedContent, "new\n")
	}
	origDest, err := os.ReadFile(dest) //nolint:gosec // test-written file
	if err != nil {
		t.Fatalf("original destination should survive: %v", err)
	}
	if string(origDest) != "old\n" {
		t.Errorf("original dest content = %q, want %q", origDest, "old\n")
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be moved away: %v", err)
	}
}

func TestSourceMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "nope.txt")
	dest := filepath.Join(dir, "dest.txt")

	out, errOut, err := run(t, "", src, dest)
	if err == nil {
		t.Fatal("expected error for missing source")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if !strings.Contains(errOut, "doesn't exist") {
		t.Errorf("stderr = %q, want it to mention the missing source", errOut)
	}
}

func TestSameSourceAndDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	writeFile(t, src, "x\n")

	_, errOut, err := run(t, "", src, src)
	if err == nil {
		t.Fatal("expected error when source and destination are the same")
	}
	if !strings.Contains(errOut, "is same") {
		t.Errorf("stderr = %q, want it to report same path", errOut)
	}
}

func TestInvalidOptionCombination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dest := filepath.Join(dir, "b.txt")
	writeFile(t, src, "x\n")

	if _, _, err := run(t, "", "-n", "-b", src, dest); err == nil {
		t.Fatal("expected error for --no-clobber with --backup")
	}
}

func TestMissingOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "")
	if err == nil {
		t.Fatal("expected error for missing operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	if errOut != "mv: missing file operand\n" {
		t.Errorf("stderr = %q, want %q", errOut, "mv: missing file operand\n")
	}
}

func TestMissingDestinationOperand(t *testing.T) {
	t.Parallel()
	out, errOut, err := run(t, "", "only-src")
	if err == nil {
		t.Fatal("expected error for missing destination operand")
	}
	if out != "" {
		t.Errorf("out = %q, want empty", out)
	}
	want := "mv: missing destination file operand after 'only-src'\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}
