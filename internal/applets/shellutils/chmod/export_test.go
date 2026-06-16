package chmod

import "os"

// ApplyModeForTest exposes the unexported applyMode for external package tests.
func ApplyModeForTest(cur os.FileMode, mode string, isDir bool) (os.FileMode, error) {
	return applyMode(cur, mode, isDir)
}

// UnwrapForTest exposes the unexported unwrap helper so the external test
// package can verify both the *os.PathError and pass-through branches.
func UnwrapForTest(err error) error {
	return unwrap(err)
}
