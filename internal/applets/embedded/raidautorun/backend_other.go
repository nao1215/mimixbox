//go:build !linux

package raidautorun

import "errors"

// errUnsupported is returned on platforms without the Linux md driver.
var errUnsupported = errors.New("RAID autorun is only supported on Linux")

// osAutoRunner is a no-op runner for non-Linux platforms.
type osAutoRunner struct{}

// Run always fails with errUnsupported.
func (osAutoRunner) Run(string) error { return errUnsupported }
