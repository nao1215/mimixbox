//go:build !linux

package makedevs

import (
	"errors"
	"os"
)

// errUnsupported is returned on platforms without mknod(2).
var errUnsupported = errors.New("device nodes can only be created on Linux")

// osNodeMaker is a no-op node maker for non-Linux platforms.
type osNodeMaker struct{}

// Mknod always fails with errUnsupported.
func (osNodeMaker) Mknod(string, byte, os.FileMode, uint32, uint32) error { return errUnsupported }
