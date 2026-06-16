//go:build unix

package find_test

import "syscall"

// mkfifo creates a FIFO (named pipe) at path on Unix-like systems.
func mkfifo(path string) error {
	return syscall.Mkfifo(path, 0o600)
}
