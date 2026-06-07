package addShell

// AddShellsForTest exposes the unexported addShells helper to the external test
// package so it can drive the file-level logic against a temp file.
func AddShellsForTest(path string, shells []string) error {
	return addShells(path, shells)
}
