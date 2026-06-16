package rm_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreserveRootRefusesSlash verifies that the default --preserve-root guard
// refuses to recursively operate on "/". The guard returns before any removal,
// so this is safe: nothing is deleted.
func TestPreserveRootRefusesSlash(t *testing.T) {
	t.Parallel()
	_, errOut, err := run(t, nil, "-r", "/")
	if err == nil {
		t.Fatal("expected a nonzero exit when removing '/' with --preserve-root")
	}
	if !strings.Contains(errOut, "it is dangerous to operate recursively on '/'") {
		t.Errorf("stderr missing danger message:\n%s", errOut)
	}
	if !strings.Contains(errOut, "use --no-preserve-root to override this failsafe") {
		t.Errorf("stderr missing override hint:\n%s", errOut)
	}
}

// TestNoPreserveRootOverridesGuard verifies --no-preserve-root disables the
// guard. We assert the danger message is NOT printed; we point rm at a private
// temp directory (never the real "/") so nothing dangerous happens.
func TestNoPreserveRootOverridesGuard(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "victim")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "f.txt"))

	_, errOut, err := run(t, nil, "-r", "--no-preserve-root", sub)
	if err != nil {
		t.Fatalf("rm -r --no-preserve-root err = %v, stderr=%s", err, errOut)
	}
	if strings.Contains(errOut, "dangerous to operate recursively") {
		t.Errorf("guard should be disabled, stderr=%s", errOut)
	}
	if exists(sub) {
		t.Errorf("%s should have been removed", sub)
	}
}

// TestPreserveRootOnlyGuardsRoot ensures the failsafe does not interfere with
// removing an ordinary directory: a non-"/" operand is removed normally.
func TestPreserveRootOnlyGuardsRoot(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "d")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(sub, "a"))

	if _, _, err := run(t, nil, "-r", sub); err != nil {
		t.Fatalf("rm -r on ordinary dir err = %v", err)
	}
	if exists(sub) {
		t.Errorf("%s should have been removed", sub)
	}
}

// TestOneFileSystemRemovesSameDevice verifies that with --one-file-system a
// tree that lives entirely on one filesystem is removed normally (the common,
// non-boundary case; we cannot mount a second filesystem in a unit test).
func TestOneFileSystemRemovesSameDevice(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	root := filepath.Join(dir, "tree")
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(nested, "leaf.txt"))
	writeFile(t, filepath.Join(root, "top.txt"))

	if _, errOut, err := run(t, nil, "-r", "--one-file-system", root); err != nil {
		t.Fatalf("rm -r --one-file-system err = %v, stderr=%s", err, errOut)
	}
	if exists(root) {
		t.Errorf("%s should have been removed entirely on a single filesystem", root)
	}
}
