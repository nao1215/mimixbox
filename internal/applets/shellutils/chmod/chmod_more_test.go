package chmod_test

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/nao1215/mimixbox/internal/applets/shellutils/chmod"
)

// TestUnwrap covers both branches of unwrap: a wrapped *os.PathError yields its
// inner error, while a plain error passes through unchanged.
func TestUnwrap(t *testing.T) {
	inner := errors.New("boom")
	pe := &os.PathError{Op: "chmod", Path: "/x", Err: inner}
	if got := chmod.UnwrapForTest(pe); got != inner {
		t.Errorf("unwrap(PathError) = %v, want %v", got, inner)
	}
	if got := chmod.UnwrapForTest(inner); got != inner {
		t.Errorf("unwrap(plain) = %v, want %v", got, inner)
	}
}

// TestApplyModeSetgidAndSticky exercises permBits/permMask/modeFromBits for the
// setgid and sticky special bits, which the existing tests do not reach.
func TestApplyModeSetgidAndSticky(t *testing.T) {
	got, err := chmod.ApplyModeForTest(0o755, "g+s", false)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSetgid == 0 {
		t.Errorf("g+s did not set setgid: %v", got)
	}

	// Round-trip a mode that already carries setuid + setgid + sticky through
	// applyMode so permBits sees all three special bits set on the input
	// (covering its setuid/setgid/sticky arms).
	cur := os.FileMode(0o755) | os.ModeSetuid | os.ModeSetgid | os.ModeSticky
	got, err = chmod.ApplyModeForTest(cur, "u+r", true)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSetuid == 0 || got&os.ModeSetgid == 0 || got&os.ModeSticky == 0 {
		t.Errorf("setuid/setgid/sticky not preserved: %v", got)
	}
}

// TestApplyModeAllSpecial covers permMask's "a+s" path (allWho sets both setuid
// and setgid) and the standalone sticky "+t".
func TestApplyModeAllSpecial(t *testing.T) {
	got, err := chmod.ApplyModeForTest(0o755, "a+s", false)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSetuid == 0 || got&os.ModeSetgid == 0 {
		t.Errorf("a+s did not set both setuid and setgid: %v", got)
	}

	got, err = chmod.ApplyModeForTest(0o755, "+t", false)
	if err != nil {
		t.Fatal(err)
	}
	if got&os.ModeSticky == 0 {
		t.Errorf("+t did not set sticky: %v", got)
	}
}

// TestApplyModeInvalidOctal covers the out-of-range octal parse error.
func TestApplyModeInvalidOctal(t *testing.T) {
	// 77777777 is octal-shaped but overflows the 32-bit parse.
	if _, err := chmod.ApplyModeForTest(0o644, "777777777777", false); err == nil {
		t.Error("expected error for overflowing octal mode")
	}
}

// TestRunInvalidModeSilent covers the silent (-f) branch of changeMode: the
// invalid-mode error is suppressed on stderr but Run still exits non-zero.
func TestRunInvalidModeSilent(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, errOut, err := run(t, "-f", "u?x", file)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if strings.Contains(errOut, "invalid mode") {
		t.Errorf("stderr = %q, want suppressed by -f", errOut)
	}
}

// TestRunMissingFileSilent covers reportAccess's silent early return: -f
// suppresses the "cannot access" diagnostic but the exit code is still non-zero.
func TestRunMissingFileSilent(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.txt")

	_, errOut, err := run(t, "-f", "644", missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if errOut != "" {
		t.Errorf("stderr = %q, want empty under -f", errOut)
	}
}

// TestRunChangesRetained covers changeMode's "retained" diagnostic: with -c and
// a no-op mode the file is unchanged, so the retained message is emitted.
func TestRunChangesRetained(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o644); err != nil {
		t.Fatal(err)
	}

	// -v with the same mode shows the "retained" line (changes==false branch).
	out, _, err := run(t, "-v", "644", file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "retained as") {
		t.Errorf("verbose stdout = %q, want retained diagnostic", out)
	}

	// -c on a no-op change should print nothing.
	out, _, err = run(t, "-c", "644", file)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("-c no-op stdout = %q, want empty", out)
	}
}

// TestRunChangesReported covers the -c "changed" branch (changes && changed).
func TestRunChangesReported(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0o600); err != nil {
		t.Fatal(err)
	}

	out, _, err := run(t, "-c", "644", file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "changed from") {
		t.Errorf("-c stdout = %q, want changed diagnostic", out)
	}
}

// TestRunRecursiveInaccessible drives changeModeRecursive's WalkDir error branch
// by recursing into a path that does not exist.
func TestRunRecursiveInaccessible(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "ghost")

	_, errOut, err := run(t, "-R", "755", missing)
	if err == nil {
		t.Fatal("expected error for missing recursive path")
	}
	if !strings.Contains(errOut, "cannot access") {
		t.Errorf("stderr = %q, want cannot-access diagnostic", errOut)
	}
}

// TestApplyModeSpecialBitsFollowWho pins which class owns each special bit, as
// in GNU chmod: setuid belongs to u, setgid to g and sticky to o. A clause
// that does not name the owning class leaves the bit alone.
func TestApplyModeSpecialBitsFollowWho(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cur  uint32
		mode string
		want uint32
	}{
		{"u= keeps sticky", 0o1644, "u=rw", 0o1644},
		{"g= keeps sticky", 0o1644, "g=r", 0o1644},
		{"o= clears sticky", 0o1644, "o=r", 0o0644},
		{"a= clears sticky", 0o1644, "a=r", 0o0444},
		{"u+t is a no-op", 0o0644, "u+t", 0o0644},
		{"g+t is a no-op", 0o0644, "g+t", 0o0644},
		{"o+t sets sticky", 0o0644, "o+t", 0o1644},
		{"+t sets sticky", 0o0644, "+t", 0o1644},
		{"u-t keeps sticky", 0o1644, "u-t", 0o1644},
		{"o= keeps setuid", 0o4755, "o=", 0o4750},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := chmod.ApplyModeForTest(modeOf(tt.cur), tt.mode, false)
			if err != nil {
				t.Fatalf("ApplyMode(%04o, %q) error = %v", tt.cur, tt.mode, err)
			}
			if bitsOf(got) != tt.want {
				t.Errorf("ApplyMode(%04o, %q) = %04o, want %04o", tt.cur, tt.mode, bitsOf(got), tt.want)
			}
		})
	}
}

// TestApplyModeOctalTooLarge checks that an octal mode above 7777 is rejected,
// as GNU chmod does, instead of silently dropping the high bits.
func TestApplyModeOctalTooLarge(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"10000", "12345", "77777"} {
		if got, err := chmod.ApplyModeForTest(0o644, mode, false); err == nil {
			t.Errorf("ApplyMode(0644, %q) = %04o, want an error", mode, bitsOf(got))
		}
	}
	got, err := chmod.ApplyModeForTest(0o644, "07777", false)
	if err != nil {
		t.Fatalf("ApplyMode(0644, \"07777\") error = %v", err)
	}
	if bitsOf(got) != 0o7777 {
		t.Errorf("ApplyMode(0644, \"07777\") = %04o, want 7777", bitsOf(got))
	}
}

// modeOf turns the 12 conventional permission bits into an os.FileMode.
func modeOf(bits uint32) os.FileMode {
	m := os.FileMode(bits & 0o777)
	if bits&0o4000 != 0 {
		m |= os.ModeSetuid
	}
	if bits&0o2000 != 0 {
		m |= os.ModeSetgid
	}
	if bits&0o1000 != 0 {
		m |= os.ModeSticky
	}
	return m
}

// bitsOf is the inverse of modeOf.
func bitsOf(m os.FileMode) uint32 {
	bits := uint32(m.Perm())
	if m&os.ModeSetuid != 0 {
		bits |= 0o4000
	}
	if m&os.ModeSetgid != 0 {
		bits |= 0o2000
	}
	if m&os.ModeSticky != 0 {
		bits |= 0o1000
	}
	return bits
}

// whoMask returns the bits a symbolic MODE may touch: the union, over its
// clauses, of the classes each clause names (u = 4700, g = 2070, o = 1007).
// A clause that names no class may touch every bit.
func whoMask(mode string) uint32 {
	var mask uint32
	for _, clause := range strings.Split(mode, ",") {
		who := clause[:len(clause)-len(strings.TrimLeft(clause, "ugoa"))]
		if who == "" || strings.Contains(who, "a") {
			return 0o7777
		}
		if strings.Contains(who, "u") {
			mask |= 0o4700
		}
		if strings.Contains(who, "g") {
			mask |= 0o2070
		}
		if strings.Contains(who, "o") {
			mask |= 0o1007
		}
	}
	return mask
}

// FuzzApplyMode checks applyMode on arbitrary modes and starting bits. It must
// never panic and must keep the file-type bits. An accepted octal mode is at
// most 7777 and becomes exactly the new mode. An accepted symbolic mode only
// changes bits that belong to the classes its clauses name, so "u=rw" cannot
// touch the sticky bit and "go-w" cannot touch setuid.
func FuzzApplyMode(f *testing.F) {
	for _, mode := range []string{
		"644", "0644", "7777", "12345", "777777777777", "u+x", "go-w", "a=r", "+x", "+X", "g+X",
		"u+x,g+r", "u+s", "g+s", "a+s", "+t", "o+t", "u=rw", "u+t", "ug=rwx,o=", "u?x", "", ",", "u=rw,",
	} {
		f.Add(mode, uint16(0o1644), false)
		f.Add(mode, uint16(0o4755), true)
	}
	f.Fuzz(func(t *testing.T, mode string, cur uint16, isDir bool) {
		before := modeOf(uint32(cur) & 0o7777)
		if isDir {
			before |= os.ModeDir
		}
		got, err := chmod.ApplyModeForTest(before, mode, isDir)
		if err != nil {
			return
		}
		if got&os.ModeType != before&os.ModeType {
			t.Fatalf("ApplyMode(%v, %q) = %v changed the file type", before, mode, got)
		}
		oldBits, newBits := bitsOf(before), bitsOf(got)
		if strings.Trim(mode, "01234567") == "" {
			want, perr := strconv.ParseUint(mode, 8, 64)
			if perr != nil || want > 0o7777 {
				t.Fatalf("ApplyMode(%v, %q) accepted an out-of-range octal mode", before, mode)
			}
			if uint64(newBits) != want {
				t.Fatalf("ApplyMode(%v, %q) = %04o, want %04o", before, mode, newBits, want)
			}
			return
		}
		if changed := oldBits ^ newBits; changed&^whoMask(mode) != 0 {
			t.Fatalf("ApplyMode(%04o, %q) = %04o changed bits %04o outside the named classes",
				oldBits, mode, newBits, changed&^whoMask(mode))
		}
	})
}
