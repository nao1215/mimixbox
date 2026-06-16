package sddf_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/sddf"
)

// chdir changes into dir for the duration of the test, restoring the previous
// working directory afterwards. sddf writes its *.sddf report relative to the
// current directory, so report-writing tests run from a temp dir.
func chdir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
}

// TestSynopsis covers the one-line description helper.
func TestSynopsis(t *testing.T) {
	if got := sddf.New().Synopsis(); got != "Search & Delete Duplicated File" {
		t.Errorf("Synopsis() = %q", got)
	}
}

// TestScanWritesReport drives the default (no -d) path: the duplicate groups are
// written to a *.sddf report. This exercises dumpToFile and decideOutputFileName.
func TestScanWritesReport(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	uniq := filepath.Join(dir, "u.txt")
	writeFile(t, a, []byte("dup-content"))
	writeFile(t, b, []byte("dup-content"))
	writeFile(t, uniq, []byte("only-me"))

	work := t.TempDir()
	chdir(t, work)

	out, _, err := run(t, "", "-o", "report", dir)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	report := filepath.Join(work, "report.sddf")
	if !strings.Contains(out, "report.sddf") {
		t.Errorf("output did not mention report file:\n%s", out)
	}
	data, rerr := os.ReadFile(report)
	if rerr != nil {
		t.Fatalf("report not written: %v", rerr)
	}
	text := string(data)
	// The report records the two duplicates and not the unique file.
	if !strings.Contains(text, a) || !strings.Contains(text, b) {
		t.Errorf("report missing duplicate paths:\n%s", text)
	}
	if strings.Contains(text, uniq) {
		t.Errorf("report unexpectedly contains the unique file:\n%s", text)
	}
	// A checksum header line is wrapped in [].
	if !strings.Contains(text, "[") || !strings.Contains(text, "]") {
		t.Errorf("report missing checksum header:\n%s", text)
	}
}

// TestDefaultOutputNameGetsExtension verifies decideOutputFileName appends the
// .sddf extension when the -o name lacks it (the default name "duplicated-file").
func TestDefaultOutputNameGetsExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a"), []byte("x"))
	writeFile(t, filepath.Join(dir, "b"), []byte("x"))

	work := t.TempDir()
	chdir(t, work)

	if _, _, err := run(t, "", dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "duplicated-file.sddf")); err != nil {
		t.Errorf("default report not written with .sddf extension: %v", err)
	}
}

// TestExplicitSddfExtensionKept verifies decideOutputFileName keeps an -o name
// that already ends in .sddf rather than doubling the extension.
func TestExplicitSddfExtensionKept(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a"), []byte("y"))
	writeFile(t, filepath.Join(dir, "b"), []byte("y"))

	work := t.TempDir()
	chdir(t, work)

	if _, _, err := run(t, "", "-o", "mine.sddf", dir); err != nil {
		t.Fatalf("Run error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "mine.sddf")); err != nil {
		t.Errorf("report not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(work, "mine.sddf.sddf")); err == nil {
		t.Error("extension was doubled to mine.sddf.sddf")
	}
}

// TestRestoreReportAndDelete writes a report, then runs sddf on that report to
// delete the duplicates it records. This exercises restoreAndDelete, restore and
// isChecksumLine end-to-end.
func TestRestoreReportAndDelete(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	writeFile(t, a, []byte("same"))
	writeFile(t, b, []byte("same"))
	writeFile(t, c, []byte("same"))

	work := t.TempDir()
	chdir(t, work)

	// 1. Scan to produce the report.
	if _, _, err := run(t, "", "-o", "rep", dir); err != nil {
		t.Fatalf("scan error = %v", err)
	}
	report := filepath.Join(work, "rep.sddf")

	// 2. Delete from the report (no -d needed: a report always deletes).
	if _, _, err := run(t, "", report); err != nil {
		t.Fatalf("restore-delete error = %v", err)
	}

	remaining := 0
	for _, p := range []string{a, b, c} {
		if _, err := os.Stat(p); err == nil {
			remaining++
		}
	}
	if remaining != 1 {
		t.Errorf("remaining files = %d, want 1 (newest kept)", remaining)
	}
}

// TestRestoreReportDryRun verifies a report can be processed with --dry-run,
// leaving the files in place.
func TestRestoreReportDryRun(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	writeFile(t, a, []byte("twin"))
	writeFile(t, b, []byte("twin"))

	work := t.TempDir()
	chdir(t, work)

	if _, _, err := run(t, "", "-o", "rep", dir); err != nil {
		t.Fatalf("scan error = %v", err)
	}
	report := filepath.Join(work, "rep.sddf")

	out, _, err := run(t, "", "--dry-run", report)
	if err != nil {
		t.Fatalf("dry-run restore error = %v", err)
	}
	if !strings.Contains(out, "Delete(DryRun)") {
		t.Errorf("dry-run output missing DryRun line:\n%s", out)
	}
	for _, p := range []string{a, b} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("dry-run removed %s: %v", p, err)
		}
	}
}

// TestNonSddfFileOperandRejected verifies a non-directory operand that does not
// carry the .sddf extension is rejected by restoreAndDelete.
func TestNonSddfFileOperandRejected(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "notes.txt")
	writeFile(t, plain, []byte("hello"))

	_, errOut, err := run(t, "", plain)
	if err == nil {
		t.Fatal("expected an error for a non-*.sddf file operand")
	}
	if !strings.Contains(errOut, "is not *.sddf") {
		t.Errorf("stderr = %q, want a format complaint", errOut)
	}
}

// TestNonexistentOperand verifies a missing operand path is reported and yields
// a failure.
func TestNonexistentOperand(t *testing.T) {
	_, errOut, err := run(t, "", "/no/such/dir/exists")
	if err == nil {
		t.Fatal("expected error for nonexistent operand")
	}
	if !strings.Contains(errOut, "sddf:") {
		t.Errorf("stderr = %q", errOut)
	}
}
