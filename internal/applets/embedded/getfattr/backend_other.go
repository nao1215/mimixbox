//go:build !linux

package getfattr

import "errors"

// errUnsupported is returned on platforms without Linux extended-attribute
// syscalls. The applet still parses and reports a deterministic error rather
// than silently succeeding.
var errUnsupported = errors.New("extended attributes are only supported on Linux")

// osBackend is a no-op backend for non-Linux platforms.
type osBackend struct{}

// List always fails with errUnsupported.
func (osBackend) List(string, bool) ([]string, error) { return nil, errUnsupported }

// Get always fails with errUnsupported.
func (osBackend) Get(string, string, bool) ([]byte, error) { return nil, errUnsupported }
