package chmod

import "os"

// ApplyModeForTest exposes the unexported applyMode for external package tests.
func ApplyModeForTest(cur os.FileMode, mode string, isDir bool) (os.FileMode, error) {
	return applyMode(cur, mode, isDir)
}
