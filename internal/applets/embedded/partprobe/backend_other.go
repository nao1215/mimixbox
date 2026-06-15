//go:build !linux

package partprobe

import "errors"

// errUnsupported is returned on platforms without BLKRRPART.
var errUnsupported = errors.New("partition-table re-read is only supported on Linux")

// osReReader is a no-op re-reader for non-Linux platforms.
type osReReader struct{}

// ReRead always fails with errUnsupported.
func (osReReader) ReRead(string) error { return errUnsupported }
