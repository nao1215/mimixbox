package find

import "os"

// readDirNames returns the names of the entries in dir.
func readDirNames(dir string) ([]string, error) {
	f, err := os.Open(dir) //nolint:gosec // operating on a user-named directory
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return f.Readdirnames(-1)
}
