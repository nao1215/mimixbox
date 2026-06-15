//go:build linux

package getfattr

import "golang.org/x/sys/unix"

// osBackend reads extended attributes through the Linux xattr syscalls.
type osBackend struct{}

// List returns the names of the extended attributes on path.
func (osBackend) List(path string, follow bool) ([]string, error) {
	list := unix.Listxattr
	if !follow {
		list = unix.Llistxattr
	}
	// Probe for the buffer size, then read.
	size, err := list(path, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}
	buf := make([]byte, size)
	n, err := list(path, buf)
	if err != nil {
		return nil, err
	}
	return splitNUL(buf[:n]), nil
}

// Get returns the value of the named extended attribute on path.
func (osBackend) Get(path, name string, follow bool) ([]byte, error) {
	get := unix.Getxattr
	if !follow {
		get = unix.Lgetxattr
	}
	size, err := get(path, name, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	n, err := get(path, name, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// splitNUL splits a NUL-separated attribute-name list into individual names.
func splitNUL(buf []byte) []string {
	var names []string
	start := 0
	for i, b := range buf {
		if b == 0 {
			if i > start {
				names = append(names, string(buf[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(buf) {
		names = append(names, string(buf[start:]))
	}
	return names
}
