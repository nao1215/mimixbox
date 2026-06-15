//go:build !linux

package watchdog

import "errors"

// errUnsupported is returned on platforms without a Linux watchdog device.
var errUnsupported = errors.New("watchdog devices are only supported on Linux")

// osPinger is a no-op pinger for non-Linux platforms.
type osPinger struct{}

// Open always fails with errUnsupported.
func (osPinger) Open(string, int) (func() error, func() error, error) {
	return nil, nil, errUnsupported
}
