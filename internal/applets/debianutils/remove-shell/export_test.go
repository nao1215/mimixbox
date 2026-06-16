package removeShell

// RemoveShellsForTest exposes the unexported removeShells helper to the external
// test package so it can drive the file-level logic against a temp file.
func RemoveShellsForTest(path string, shells []string) error {
	return removeShells(path, shells)
}

// ReadShellsForTest exposes the unexported readShells helper to the external
// test package.
func ReadShellsForTest(path string) ([]string, error) {
	return readShells(path)
}
