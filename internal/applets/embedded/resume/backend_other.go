//go:build !linux

package resume

import "errors"

// errUnsupported is returned on platforms without Linux hibernation support.
var errUnsupported = errors.New("resume from hibernation is only supported on Linux")

// osResolver is a no-op resolver for non-Linux platforms.
type osResolver struct{}

// DevNumber always fails with errUnsupported.
func (osResolver) DevNumber(string) (string, error) { return "", errUnsupported }
