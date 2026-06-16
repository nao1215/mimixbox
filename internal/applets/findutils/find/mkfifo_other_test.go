//go:build !unix

package find_test

import "errors"

// mkfifo is unsupported on non-Unix platforms; the FIFO test skips when it
// returns an error.
func mkfifo(path string) error {
	return errors.New("mkfifo not supported on this platform")
}
