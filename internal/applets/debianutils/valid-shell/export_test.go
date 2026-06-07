package validShell

import "io"

// ValidateShellsForTest exposes the unexported validateShells helper to the
// external test package so it can drive the file-level logic against a temp
// file.
func ValidateShellsForTest(path string, out io.Writer) (bool, error) {
	return validateShells(path, out)
}
