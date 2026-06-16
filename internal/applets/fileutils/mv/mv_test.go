package mv_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// TestTargetDirectory checks that -t DIR moves every source into DIR
// (destination-first form), keeping each source's base name.
func TestTargetDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeFile(t, a, "A")
	writeFile(t, b, "B")
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "", "-t", destDir, a, b)
	if err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	for name, want := range map[string]string{"a.txt": "A", "b.txt": "B"} {
		got, rerr := os.ReadFile(filepath.Join(destDir, name)) //nolint:gosec // test path
		if rerr != nil {
			t.Errorf("%s not moved into target dir: %v", name, rerr)
			continue
		}
		if string(got) != want {
			t.Errorf("%s content = %q, want %q", name, got, want)
		}
	}
	// Sources must be gone.
	for _, p := range []string{a, b} {
		if _, serr := os.Stat(p); !os.IsNotExist(serr) {
			t.Errorf("source %s still exists after move", p)
		}
	}
}

// TestNoTargetDirectory checks that -T treats the destination as a normal file
// even when a directory of the same name already exists: the move must fail
// rather than descend into the directory.
func TestNoTargetDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	writeFile(t, src, "S")
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "", "-T", src, destDir)
	if err == nil {
		t.Fatal("expected error: -T must not descend into an existing directory")
	}
	// destDir must remain a directory and not contain the source basename.
	if info, serr := os.Stat(destDir); serr != nil || !info.IsDir() {
		t.Errorf("dest no longer a directory: info=%v err=%v", info, serr)
	}
	if _, serr := os.Stat(filepath.Join(destDir, "src.txt")); serr == nil {
		t.Error("-T must not move the source inside the directory")
	}
}

// TestNoTargetDirectoryRejectsExtraOperand checks -T rejects the multi-source
// form.
func TestNoTargetDirectoryRejectsExtraOperand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeFile(t, a, "A")
	writeFile(t, b, "B")
	destDir := filepath.Join(dir, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, errOut, err := run(t, "", "-T", a, b, destDir)
	if err == nil {
		t.Fatal("expected error for extra operand with -T")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q, want extra-operand message", errOut)
	}
}

// TestUpdateSkipsNewerDestination checks that -u does not overwrite a
// destination that is newer than the source, but does move when the
// destination is older or missing.
func TestUpdateSkipsNewerDestination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	writeFile(t, src, "source")
	writeFile(t, dest, "newer-dest")

	// Make the destination strictly newer than the source.
	old := time.Now().Add(-1 * time.Hour)
	now := time.Now()
	if err := os.Chtimes(src, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dest, now, now); err != nil {
		t.Fatal(err)
	}

	// -u must skip: destination is newer, and the source must remain.
	if _, errOut, err := run(t, "", "-u", src, dest); err != nil {
		t.Fatalf("Run -u (skip) error = %v (stderr=%q)", err, errOut)
	}
	got, _ := os.ReadFile(dest) //nolint:gosec // test path
	if string(got) != "newer-dest" {
		t.Errorf("dest content = %q, want it preserved as %q", got, "newer-dest")
	}
	if _, serr := os.Stat(src); serr != nil {
		t.Errorf("source must remain when -u skips: %v", serr)
	}

	// Now make the source newer; -u must move and overwrite the destination.
	newer := time.Now().Add(1 * time.Hour)
	if err := os.Chtimes(src, newer, newer); err != nil {
		t.Fatal(err)
	}
	if _, errOut, err := run(t, "", "-u", src, dest); err != nil {
		t.Fatalf("Run -u (move) error = %v (stderr=%q)", err, errOut)
	}
	got, _ = os.ReadFile(dest) //nolint:gosec // test path
	if string(got) != "source" {
		t.Errorf("dest content = %q, want overwrite to %q", got, "source")
	}
	if _, serr := os.Stat(src); !os.IsNotExist(serr) {
		t.Error("source must be gone after -u moves")
	}
}

// TestUpdateMovesWhenDestMissing checks -u always moves when the destination
// does not yet exist.
func TestUpdateMovesWhenDestMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dest := filepath.Join(dir, "dest.txt")
	writeFile(t, src, "data")

	if _, errOut, err := run(t, "", "-u", src, dest); err != nil {
		t.Fatalf("Run error = %v (stderr=%q)", err, errOut)
	}
	got, rerr := os.ReadFile(dest) //nolint:gosec // test path
	if rerr != nil {
		t.Fatalf("dest not created: %v", rerr)
	}
	if string(got) != "data" {
		t.Errorf("dest content = %q, want %q", got, "data")
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
