package chmod_test

import (
	"errors"
	"os"
	"path/filepath"
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
