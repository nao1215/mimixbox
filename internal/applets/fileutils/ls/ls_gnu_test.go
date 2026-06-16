package ls

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// gnuFixture builds a directory with a subdir, an executable, a symlink, files
// of distinct sizes/mtimes, and pattern-matching names (*.log, tmp*).
func gnuFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "small.txt"), 10, 0o644)
	mustWrite(t, filepath.Join(dir, "big.txt"), 5000, 0o644)
	mustWrite(t, filepath.Join(dir, "run.sh"), 20, 0o755)
	mustWrite(t, filepath.Join(dir, "a.log"), 1, 0o644)
	mustWrite(t, filepath.Join(dir, "tmpfile"), 1, 0o644)
	if err := os.MkdirAll(filepath.Join(dir, "adir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("small.txt", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	// Distinct mtimes for time sorting.
	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now().Add(-1 * time.Hour)
	_ = os.Chtimes(filepath.Join(dir, "small.txt"), old, old)
	_ = os.Chtimes(filepath.Join(dir, "big.txt"), recent, recent)
	return dir
}

func mustWrite(t *testing.T, path string, n int, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Repeat("x", n)), mode); err != nil {
		t.Fatal(err)
	}
}

// ---- #722 color ----------------------------------------------------------

func TestParseColorMode(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"always": "always", "auto": "auto", "never": "never",
		"yes": "always", "no": "never", "tty": "auto", "": "never",
	}
	for in, want := range cases {
		got, err := parseColorMode(in)
		if err != nil || got != want {
			t.Errorf("parseColorMode(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := parseColorMode("bogus"); err == nil {
		t.Error("parseColorMode(bogus) should error")
	}
}

func TestColorAlways(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "--color=always", gnuFixture(t))
	if !strings.Contains(out, colorDir+"adir"+colorReset) {
		t.Errorf("dir should be colored bold blue: %q", out)
	}
	if !strings.Contains(out, colorExec+"run.sh"+colorReset) {
		t.Errorf("exec should be colored bold green: %q", out)
	}
	if !strings.Contains(out, colorSymlink+"link"+colorReset) {
		t.Errorf("symlink should be colored bold cyan: %q", out)
	}
}

func TestColorNeverAndAuto(t *testing.T) {
	t.Parallel()
	for _, when := range []string{"never", "auto"} {
		out, _ := run(t, "--color="+when, gnuFixture(t))
		if strings.Contains(out, "\x1b[") {
			t.Errorf("--color=%s should emit no escapes (non-tty): %q", when, out)
		}
	}
}

// ---- #723 indicators -----------------------------------------------------

func TestResolveIndicator(t *testing.T) {
	t.Parallel()
	if got := resolveIndicator(true, false, ""); got != indicatorClassify {
		t.Errorf("-F => classify, got %v", got)
	}
	if got := resolveIndicator(false, true, ""); got != indicatorFileType {
		t.Errorf("--file-type => file-type, got %v", got)
	}
	if got := resolveIndicator(false, false, "slash"); got != indicatorSlash {
		t.Errorf("--indicator-style=slash, got %v", got)
	}
	if got := resolveIndicator(true, false, "none"); got != indicatorNone {
		t.Errorf("explicit none should win over -F, got %v", got)
	}
}

func TestClassifyIndicators(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "-F", gnuFixture(t))
	for _, want := range []string{"adir/", "run.sh*", "link@"} {
		if !strings.Contains(out, want) {
			t.Errorf("-F missing %q in %q", want, out)
		}
	}
}

func TestFileTypeIndicators(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "--file-type", gnuFixture(t))
	if !strings.Contains(out, "adir/") {
		t.Errorf("--file-type should mark dirs: %q", out)
	}
	if strings.Contains(out, "run.sh*") {
		t.Errorf("--file-type must NOT add * for executables: %q", out)
	}
	if !strings.Contains(out, "link@") {
		t.Errorf("--file-type should mark symlinks: %q", out)
	}
}

func TestIndicatorStyleSlash(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "--indicator-style=slash", gnuFixture(t))
	if !strings.Contains(out, "adir/") {
		t.Errorf("slash should mark dirs: %q", out)
	}
	if strings.Contains(out, "run.sh*") || strings.Contains(out, "link@") {
		t.Errorf("slash should only mark dirs: %q", out)
	}
}

// ---- #724 sorting --------------------------------------------------------

func TestParseSort(t *testing.T) {
	t.Parallel()
	cases := map[string]sortKey{
		"name": sortName, "none": sortNone, "size": sortSize,
		"time": sortTime, "version": sortVersion, "extension": sortExtension,
	}
	for in, want := range cases {
		got, err := parseSort(in, false)
		if err != nil || got != want {
			t.Errorf("parseSort(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if got, _ := parseSort("name", true); got != sortNone {
		t.Errorf("-U should force sortNone, got %v", got)
	}
	if _, err := parseSort("bogus", false); err == nil {
		t.Error("parseSort(bogus) should error")
	}
}

// entriesFor builds entry values (with cached metadata) for the named files in
// dir, mirroring how the listing pipeline populates them.
func entriesFor(dir string, names ...string) []entry {
	es := make([]entry, 0, len(names))
	for _, n := range names {
		es = append(es, newEntry(dir, n))
	}
	return es
}

func entryNames(es []entry) []string {
	names := make([]string, len(es))
	for i, e := range es {
		names[i] = e.name
	}
	return names
}

func TestSortSize(t *testing.T) {
	t.Parallel()
	dir := gnuFixture(t)
	es := entriesFor(dir, "small.txt", "big.txt", "a.log")
	sortEntries(es, options{sortBy: sortSize})
	if es[0].name != "big.txt" {
		t.Errorf("size sort should put big.txt first: %v", entryNames(es))
	}
}

func TestSortTime(t *testing.T) {
	t.Parallel()
	dir := gnuFixture(t)
	es := entriesFor(dir, "small.txt", "big.txt")
	sortEntries(es, options{sortBy: sortTime, timeBy: timeMtime})
	// big.txt is more recent, so it sorts first.
	if es[0].name != "big.txt" {
		t.Errorf("time sort newest-first wrong: %v", entryNames(es))
	}
}

func TestSortVersion(t *testing.T) {
	t.Parallel()
	es := entriesFor("", "file10", "file2", "file1")
	sortEntries(es, options{sortBy: sortVersion})
	want := []string{"file1", "file2", "file10"}
	got := entryNames(es)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("version sort = %v, want %v", got, want)
		}
	}
}

func TestSortExtension(t *testing.T) {
	t.Parallel()
	es := entriesFor("", "b.txt", "a.log", "c.txt")
	sortEntries(es, options{sortBy: sortExtension})
	if es[0].name != "a.log" {
		t.Errorf(".log should sort before .txt: %v", entryNames(es))
	}
}

func TestGroupDirectoriesFirst(t *testing.T) {
	t.Parallel()
	dir := gnuFixture(t)
	out, _ := run(t, "--group-directories-first", dir)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if lines[0] != "adir" {
		t.Errorf("group-dirs-first should list adir first: %v", lines)
	}
}

func TestVersionCompare(t *testing.T) {
	t.Parallel()
	if versionCompare("file2", "file10") >= 0 {
		t.Error("file2 < file10")
	}
	if versionCompare("a", "a") != 0 {
		t.Error("equal strings")
	}
	if versionCompare("b", "a") <= 0 {
		t.Error("b > a")
	}
}

// ---- #725 hide / ignore --------------------------------------------------

func TestIgnoreAlwaysFilters(t *testing.T) {
	t.Parallel()
	out, _ := run(t, "--ignore=*.log", gnuFixture(t))
	if strings.Contains(out, "a.log") {
		t.Errorf("--ignore should drop a.log: %q", out)
	}
	out2, _ := run(t, "-a", "--ignore=*.log", gnuFixture(t))
	if strings.Contains(out2, "a.log") {
		t.Errorf("--ignore should drop even with -a: %q", out2)
	}
}

func TestHideOverriddenByAll(t *testing.T) {
	t.Parallel()
	// Without -a/-A, hide applies.
	out, _ := run(t, "--hide=tmp*", gnuFixture(t))
	if strings.Contains(out, "tmpfile") {
		t.Errorf("--hide should drop tmpfile: %q", out)
	}
	// With -a, hide is ignored.
	out2, _ := run(t, "-a", "--hide=tmp*", gnuFixture(t))
	if !strings.Contains(out2, "tmpfile") {
		t.Errorf("--hide must be overridden by -a: %q", out2)
	}
}

func TestFiltered(t *testing.T) {
	t.Parallel()
	if !filtered("a.log", options{ignore: "*.log"}) {
		t.Error("ignore *.log should filter a.log")
	}
	if filtered("a.log", options{ignore: "*.txt"}) {
		t.Error("ignore *.txt should not filter a.log")
	}
	if !filtered("tmpx", options{hide: "tmp*"}) {
		t.Error("hide tmp* should filter tmpx")
	}
	if filtered("tmpx", options{hide: "tmp*", all: true}) {
		t.Error("hide must be disabled when all is set")
	}
}

// ---- #726 inode / block-size --------------------------------------------

func TestInode(t *testing.T) {
	t.Parallel()
	dir := gnuFixture(t)
	out, _ := run(t, "-i", "--ignore=*", dir, filepath.Join(dir, "small.txt"))
	// -i prints inode before the name; ensure the line starts with a number.
	got := newEntry(dir, "small.txt").inode()
	if got == 0 {
		t.Skip("inode not available on this platform")
	}
	if !strings.Contains(out, "small.txt") {
		t.Fatalf("output missing small.txt: %q", out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.Contains(line, "small.txt") {
			if line[0] < '0' || line[0] > '9' {
				t.Errorf("-i line should start with inode number: %q", line)
			}
		}
	}
}

func TestParseSize(t *testing.T) {
	t.Parallel()
	cases := map[string]int64{
		"1024": 1024, "K": 1024, "1K": 1024, "2K": 2048,
		"1M": 1024 * 1024, "1KB": 1000,
	}
	for in, want := range cases {
		got, err := parseSize(in)
		if err != nil || got != want {
			t.Errorf("parseSize(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	if _, err := parseSize("bogus"); err == nil {
		t.Error("parseSize(bogus) should error")
	}
}

func TestResolveBlockSize(t *testing.T) {
	t.Parallel()
	if got, _ := resolveBlockSize("", false); got != 0 {
		t.Errorf("default block size = %d, want 0", got)
	}
	if got, _ := resolveBlockSize("", true); got != 1024 {
		t.Errorf("-k block size = %d, want 1024", got)
	}
	if got, _ := resolveBlockSize("512", false); got != 512 {
		t.Errorf("--block-size=512 = %d", got)
	}
	if _, err := resolveBlockSize("bad", false); err == nil {
		t.Error("--block-size=bad should error")
	}
}

func TestSizeStringBlockSize(t *testing.T) {
	t.Parallel()
	// 5000 bytes in 1024-byte blocks rounds up to 5.
	if got := sizeString(5000, false, 1024); got != "5" {
		t.Errorf("sizeString(5000,1024) = %q, want 5", got)
	}
	if got := sizeString(1024, false, 1024); got != "1" {
		t.Errorf("sizeString(1024,1024) = %q, want 1", got)
	}
	if got := sizeString(0, false, 1024); got != "0" {
		t.Errorf("sizeString(0,1024) = %q, want 0", got)
	}
}

func TestBlockSizeInLong(t *testing.T) {
	t.Parallel()
	dir := gnuFixture(t)
	out, _ := run(t, "-l", "-k", "--ignore=*", dir, filepath.Join(dir, "big.txt"))
	// big.txt is 5000 bytes => 5 blocks of 1024.
	line := firstMatching(out, "big.txt")
	if !strings.Contains(line, " 5 ") {
		t.Errorf("-l -k should show 5 blocks for big.txt: %q", line)
	}
}

func firstMatching(out, sub string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, sub) {
			return line
		}
	}
	return ""
}
