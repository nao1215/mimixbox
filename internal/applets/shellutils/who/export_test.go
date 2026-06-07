package who

// SetUtmpFileForTest points the applet at path and returns a function that
// restores the previous value, so external tests can supply a fixture utmp file.
func SetUtmpFileForTest(path string) (restore func()) {
	prev := utmpFile
	utmpFile = path
	return func() { utmpFile = prev }
}
