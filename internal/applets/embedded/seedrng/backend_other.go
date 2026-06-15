//go:build !linux

package seedrng

import "errors"

// errUnsupported is returned on platforms without the Linux RNDADDENTROPY ioctl.
var errUnsupported = errors.New("seeding the kernel RNG is only supported on Linux")

// osSeeder is a no-op seeder for non-Linux platforms.
type osSeeder struct{}

// Credit always fails with errUnsupported.
func (osSeeder) Credit([]byte, bool) error { return errUnsupported }
