package touch

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func atimeOf(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skip("syscall.Stat_t unavailable")
	}
	return time.Unix(st.Atim.Sec, st.Atim.Nsec)
}

// TestReferenceCopiesMtime verifies --reference copies the reference file's
// modification time onto the target.
func TestReferenceCopiesMtime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	ref := filepath.Join(dir, "ref")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(ref, []byte("r"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("d"), 0o600); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2001, 2, 3, 4, 5, 6, 0, time.Local)
	if err := os.Chtimes(ref, want, want); err != nil {
		t.Fatal(err)
	}

	rt, err := referenceTimes(ref)
	if err != nil {
		t.Fatalf("referenceTimes err = %v", err)
	}
	if err := touch(dst, options{useTimes: true, atime: rt.atime, mtime: rt.mtime}); err != nil {
		t.Fatalf("touch --reference err = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(want) {
		t.Errorf("mtime = %v, want %v", info.ModTime(), want)
	}
}

// TestDateSetsKnownTime verifies --date sets both timestamps to a parsed value.
func TestDateSetsKnownTime(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dst := filepath.Join(dir, "f")
	if err := os.WriteFile(dst, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	when, err := parseDate("2020-06-15 12:34:56")
	if err != nil {
		t.Fatalf("parseDate err = %v", err)
	}
	if err := touch(dst, options{useTimes: true, atime: when, mtime: when}); err != nil {
		t.Fatalf("touch --date err = %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(when) {
		t.Errorf("mtime = %v, want %v", info.ModTime(), when)
	}
}

// TestParseDateUnix verifies the @UNIX form.
func TestParseDateUnix(t *testing.T) {
	t.Parallel()
	got, err := parseDate("@1000000000")
	if err != nil {
		t.Fatalf("parseDate @ err = %v", err)
	}
	if !got.Equal(time.Unix(1000000000, 0)) {
		t.Errorf("got %v, want %v", got, time.Unix(1000000000, 0))
	}
}

// TestParseDateInvalid verifies a bad string is rejected.
func TestParseDateInvalid(t *testing.T) {
	t.Parallel()
	if _, err := parseDate("not a date"); err == nil {
		t.Fatal("expected an error for an unparseable date")
	}
}

// TestTimeAtimeSetsAccessOnly verifies --time=atime (modelled here as
// accessOnly) advances the access time while leaving the modification time at
// its old value.
func TestTimeAtimeSetsAccessOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dst := filepath.Join(dir, "f")
	if err := os.WriteFile(dst, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-5 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(dst, old, old); err != nil {
		t.Fatal(err)
	}

	newer := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	if err := touch(dst, options{accessOnly: true, useTimes: true, atime: newer, mtime: newer}); err != nil {
		t.Fatalf("touch --time=atime err = %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	// mtime must be preserved at old.
	if d := info.ModTime().Sub(old); d < -time.Second || d > time.Second {
		t.Errorf("mtime moved by %v, want unchanged", d)
	}
	// atime must have advanced to newer.
	if d := atimeOf(t, dst).Sub(newer); d < -time.Second || d > time.Second {
		t.Errorf("atime = %v, want ~%v", atimeOf(t, dst), newer)
	}
}

// TestNoDereferenceTouchesLinkNotTarget verifies -h changes the symlink's own
// timestamps and leaves the target's modification time untouched.
func TestNoDereferenceTouchesLinkNotTarget(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("t"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-6 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(target, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	when := time.Now().Add(3 * time.Hour).Truncate(time.Second)
	if err := touch(link, options{noDereference: true, useTimes: true, atime: when, mtime: when}); err != nil {
		t.Fatalf("touch -h err = %v", err)
	}

	// Target's mtime must be unchanged (still old).
	ti, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if d := ti.ModTime().Sub(old); d < -time.Second || d > time.Second {
		t.Errorf("target mtime moved by %v, want unchanged with -h", d)
	}
	// The link's own mtime should be the new value.
	li, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if d := li.ModTime().Sub(when); d < -time.Second || d > time.Second {
		t.Errorf("link mtime = %v, want ~%v", li.ModTime(), when)
	}
}
