package halt

import "testing"

// SetRebootFnForTest replaces the dangerous reboot syscall with fn for the
// duration of a test, restoring the original afterwards. This keeps the real
// syscall.Reboot out of every test.
func SetRebootFnForTest(t *testing.T, fn func(int) error) {
	t.Helper()
	orig := rebootFn
	rebootFn = fn
	t.Cleanup(func() { rebootFn = orig })
}

// SetIsRootForTest replaces the root check with fn for the duration of a test,
// so Run can be exercised without real root privileges.
func SetIsRootForTest(t *testing.T, fn func() bool) {
	t.Helper()
	orig := isRoot
	isRoot = fn
	t.Cleanup(func() { isRoot = orig })
}
