//go:build !linux

package i2c

import "errors"

// errUnsupported is returned on platforms without the Linux I2C subsystem.
var errUnsupported = errors.New("I2C access is only supported on Linux")

// osBackend is a no-op backend for non-Linux platforms.
type osBackend struct{}

// ReadReg always fails with errUnsupported.
func (osBackend) ReadReg(int, int, int) (byte, error) { return 0, errUnsupported }

// WriteReg always fails with errUnsupported.
func (osBackend) WriteReg(int, int, int, byte) error { return errUnsupported }

// Detect always fails with errUnsupported.
func (osBackend) Detect(int, int, int) ([]int, error) { return nil, errUnsupported }
