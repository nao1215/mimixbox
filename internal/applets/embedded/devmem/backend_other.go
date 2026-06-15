//go:build !linux

package devmem

import "errors"

// errUnsupported is returned on platforms without /dev/mem.
var errUnsupported = errors.New("physical memory access is only supported on Linux")

// osBackend is a no-op backend for non-Linux platforms.
type osBackend struct{}

// Read always fails with errUnsupported.
func (osBackend) Read(uint64, int) (uint64, error) { return 0, errUnsupported }

// Write always fails with errUnsupported.
func (osBackend) Write(uint64, int, uint64) error { return errUnsupported }
