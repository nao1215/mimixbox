//go:build linux

package setfattr

import "golang.org/x/sys/unix"

// osBackend writes extended attributes through the Linux xattr syscalls.
type osBackend struct{}

// Set stores value under name on path.
func (osBackend) Set(path, name string, value []byte, follow bool) error {
	if follow {
		return unix.Setxattr(path, name, value, 0)
	}
	return unix.Lsetxattr(path, name, value, 0)
}

// Remove deletes the named attribute from path.
func (osBackend) Remove(path, name string, follow bool) error {
	if follow {
		return unix.Removexattr(path, name)
	}
	return unix.Lremovexattr(path, name)
}
