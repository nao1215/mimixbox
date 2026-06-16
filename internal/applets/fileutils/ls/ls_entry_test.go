package ls

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/nao1215/mimixbox/internal/command"
)

// richFixture builds a directory holding one of each entry kind the listing
// pipeline decorates differently: a regular file, a directory, a symlink, and
// an executable. It returns the directory path.
func richFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "reg.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "adir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Point at a name that does not collide with any listed entry, so line
	// matching by name fragment is unambiguous.
	if err := os.Symlink("elsewhere", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	return dir
}

// lineFor returns the single output line that ends with the given display name
// fragment, to keep assertions stable against owner/group/time variation.
func lineFor(t *testing.T, out, frag string) string {
	t.Helper()
	for _, ln := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.Contains(ln, frag) {
			return ln
		}
	}
	t.Fatalf("no line containing %q in:\n%s", frag, out)
	return ""
}

// TestCombinedLongClassifyInode locks the combined -l -F -i rendering across a
// regular file, directory, symlink, and executable, with color disabled (the
// default). This is the behavior the entry-model refactor must keep byte-stable.
func TestCombinedLongClassifyInode(t *testing.T) {
	t.Parallel()
	dir := richFixture(t)
	out, err := run(t, "-l", "-F", "-i", dir)
	if err != nil {
		t.Fatalf("ls -lFi error = %v", err)
	}
	// No ANSI escapes when color is off.
	if strings.Contains(out, "\x1b[") {
		t.Errorf("color should be off by default, got escapes: %q", out)
	}

	// Each line is prefixed by the real inode number from Lstat.
	for _, name := range []string{"reg.txt", "adir", "run.sh", "link"} {
		info, err := os.Lstat(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			t.Skip("no syscall.Stat_t on this platform")
		}
		ln := lineFor(t, out, name)
		if !strings.HasPrefix(ln, strconv.FormatUint(st.Ino, 10)+" ") {
			t.Errorf("line for %s = %q, want inode %d prefix", name, ln, st.Ino)
		}
	}

	// Directory: mode starts with 'd', name has a trailing '/'.
	if ln := lineFor(t, out, "adir"); !strings.Contains(ln, " adir/") || !strings.Contains(ln, " d") {
		t.Errorf("dir line = %q, want 'd' mode and adir/ suffix", ln)
	}
	// Executable: classify appends '*'.
	if ln := lineFor(t, out, "run.sh"); !strings.HasSuffix(ln, "run.sh*") {
		t.Errorf("exec line = %q, want run.sh* suffix", ln)
	}
	// Symlink: mode starts with 'l' and the target is shown.
	if ln := lineFor(t, out, "link"); !strings.Contains(ln, "link -> elsewhere") || !strings.Contains(ln, " l") {
		t.Errorf("symlink line = %q, want 'l' mode and 'link -> elsewhere'", ln)
	}
}

// TestColorDisabledPlain proves the non-long path emits no ANSI escapes when
// color is off, for every entry kind.
func TestColorDisabledPlain(t *testing.T) {
	t.Parallel()
	out, err := run(t, "-F", richFixture(t))
	if err != nil {
		t.Fatalf("ls -F error = %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Errorf("unexpected ANSI escapes with color off: %q", out)
	}
	for _, want := range []string{"adir/", "link@", "run.sh*", "reg.txt"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %q", want, out)
		}
	}
}

// TestColorEnabledWrapsTypes locks the color rendering for each type so the
// refactor keeps the same escapes around dir/exec/symlink names.
func TestColorEnabledWrapsTypes(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--color=always", richFixture(t))
	if err != nil {
		t.Fatalf("ls --color=always error = %v", err)
	}
	if !strings.Contains(out, colorDir+"adir"+colorReset) {
		t.Errorf("dir not colored: %q", out)
	}
	if !strings.Contains(out, colorExec+"run.sh"+colorReset) {
		t.Errorf("exec not colored: %q", out)
	}
	if !strings.Contains(out, colorSymlink+"link"+colorReset) {
		t.Errorf("symlink not colored: %q", out)
	}
}

// sortFixture writes files with controlled sizes, mtimes, names and a directory
// so the sort-key + group-directories-first combinations are deterministic.
func sortFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, size int, age time.Duration) {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(strings.Repeat("x", size)), 0o644); err != nil {
			t.Fatal(err)
		}
		when := time.Now().Add(-age)
		if err := os.Chtimes(p, when, when); err != nil {
			t.Fatal(err)
		}
	}
	write("big.log", 300, 3*time.Hour)   // largest, oldest
	write("mid.txt", 200, 2*time.Hour)   // middle
	write("small.dat", 100, 1*time.Hour) // smallest, newest
	if err := os.MkdirAll(filepath.Join(dir, "zdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func names(out string) []string {
	return strings.Split(strings.TrimRight(out, "\n"), "\n")
}

// TestSortSizeGroupDirs locks --sort=size with --group-directories-first: the
// directory comes first, then files largest-first.
func TestSortSizeGroupDirs(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--sort=size", "--group-directories-first", sortFixture(t))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	want := []string{"zdir", "big.log", "mid.txt", "small.dat"}
	if got := names(out); !equalSlice(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

// TestSortTimeGroupDirs locks --sort=time with --group-directories-first:
// directory first, then files newest-first.
func TestSortTimeGroupDirs(t *testing.T) {
	t.Parallel()
	out, err := run(t, "--sort=time", "--group-directories-first", sortFixture(t))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	want := []string{"zdir", "small.dat", "mid.txt", "big.log"}
	if got := names(out); !equalSlice(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

// TestSortVersionGroupDirs locks --sort=version with --group-directories-first.
func TestSortVersionGroupDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, n := range []string{"f2", "f10", "f1"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "d1"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "--sort=version", "--group-directories-first", dir)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	want := []string{"d1", "f1", "f2", "f10"}
	if got := names(out); !equalSlice(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

// TestSortExtensionGroupDirs locks --sort=extension with
// --group-directories-first.
func TestSortExtensionGroupDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, n := range []string{"b.go", "a.txt", "c.go"} {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "zd"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "--sort=extension", "--group-directories-first", dir)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	// Directory first; then files by extension (.go before .txt), name as tie.
	want := []string{"zd", "b.go", "c.go", "a.txt"}
	if got := names(out); !equalSlice(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// BenchmarkListManyEntriesLong measures listing a directory with many entries in
// long+inode form, where the old code restatted each entry many times. It lets
// the entry-model refactor demonstrate fewer duplicate metadata lookups without
// changing output.
func BenchmarkListManyEntriesLong(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 500; i++ {
		p := filepath.Join(dir, "file"+strconv.Itoa(i)+".txt")
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			b.Fatal(err)
		}
	}
	stdio := command.IO{In: strings.NewReader(""), Out: io.Discard, Err: io.Discard}
	cmd := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cmd.Run(context.Background(), stdio, []string{"-l", "-i", "--sort=size", dir}); err != nil {
			b.Fatal(err)
		}
	}
}
