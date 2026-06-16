package serial_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/serial"
)

// TestSynopsis covers the one-line description helper.
func TestSynopsis(t *testing.T) {
	if got := serial.New().Synopsis(); got != "Rename the file to the name with a serial number" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestNameWithDirectoryCreatesDir drives makeDirIfNeeded: when --name carries a
// directory component that does not yet exist, the destination directory is
// created and the renamed files are placed in it. With --suffix the serial
// number follows the chosen base name.
func TestNameWithDirectoryCreatesDir(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt", "b.txt")
	destDir := filepath.Join(dir, "out")
	namePath := filepath.Join(destDir, "img")

	if _, _, err := run(t, "--suffix", "--name", namePath, dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := listDir(t, destDir)
	want := []string{"img_0.txt", "img_1.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("files in created dir = %v, want %v", got, want)
	}
}

// TestExistingTargetWithoutForceFails drives dieIfExistSameNameFile: a rename
// target that already exists is refused unless --force is given. The pre-created
// collision file lives in a separate output directory (via --name) so it is not
// itself picked up as a source.
func TestExistingTargetWithoutForceFails(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt")
	outDir := t.TempDir()
	// The target for a.txt with --suffix --name=outDir/img is img_0.txt.
	if err := os.WriteFile(filepath.Join(outDir, "img_0.txt"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := run(t, "--suffix", "--name", filepath.Join(outDir, "img"), dir)
	if err == nil {
		t.Fatal("expected failure: target name already exists")
	}
}

// TestForceOverwritesExistingCopyTarget drives the copyFiles overwrite branch:
// with --keep and --force, an existing destination is removed before being
// re-created as a hard link to the source.
func TestForceOverwritesExistingCopyTarget(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt")
	outDir := t.TempDir()
	target := filepath.Join(outDir, "img_0.txt")
	// Pre-create the copy target with stale content.
	if err := os.WriteFile(target, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, _, err := run(t, "--keep", "--force", "--suffix", "--name", filepath.Join(outDir, "img"), dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	// The original a.txt is preserved (keep), and the target now mirrors it.
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "a.txt" {
		t.Errorf("overwritten copy content = %q, want %q", string(got), "a.txt")
	}
	if _, err := os.Stat(filepath.Join(dir, "a.txt")); err != nil {
		t.Errorf("original should be kept: %v", err)
	}
}

// TestExtensionlessFilePrefixed drives baseNameWithoutExt's no-extension branch:
// a file with no extension keeps its whole name as the base.
func TestExtensionlessFilePrefixed(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "README", "LICENSE")

	if _, _, err := run(t, dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got := listDir(t, dir)
	want := []string{"0_LICENSE", "1_README"}
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("renamed extensionless files = %v, want %v", got, want)
	}
}

// TestInvalidNameDirectoryOnly verifies a --name that is only a directory (ends
// with a slash) is rejected.
func TestInvalidNameDirectoryOnly(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt")

	_, errOut, err := run(t, "--name", "subdir/", dir)
	if err == nil {
		t.Fatal("expected failure for a directory-only --name")
	}
	if !strings.Contains(errOut, "invalid --name") {
		t.Errorf("stderr = %q", errOut)
	}
}

// TestExtraOperandRejected verifies a second positional operand is an error.
func TestExtraOperandRejected(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, "a.txt")

	_, errOut, err := run(t, dir, dir)
	if err == nil {
		t.Fatal("expected failure for an extra operand")
	}
	if !strings.Contains(errOut, "extra operand") {
		t.Errorf("stderr = %q", errOut)
	}
}
