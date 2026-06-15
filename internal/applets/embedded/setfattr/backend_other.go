//go:build !linux

package setfattr

import "errors"

// errUnsupported is returned on platforms without Linux extended-attribute
// syscalls. The applet still parses and reports a deterministic error.
var errUnsupported = errors.New("extended attributes are only supported on Linux")

// osBackend is a no-op backend for non-Linux platforms.
type osBackend struct{}

// Set always fails with errUnsupported.
func (osBackend) Set(string, string, []byte, bool) error { return errUnsupported }

// Remove always fails with errUnsupported.
func (osBackend) Remove(string, string, bool) error { return errUnsupported }
