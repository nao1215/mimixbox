//go:build linux

package makedevs

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// osNodeMaker creates device nodes with mknod(2). Without the CAP_MKNOD
// capability the kernel returns EPERM, which is surfaced as a documented error
// rather than a silent skip.
type osNodeMaker struct{}

// Mknod creates a special file at path.
func (osNodeMaker) Mknod(path string, kind byte, mode os.FileMode, major, minor uint32) error {
	var modeBits uint32
	switch kind {
	case 'c':
		modeBits = unix.S_IFCHR
	case 'b':
		modeBits = unix.S_IFBLK
	case 'p':
		modeBits = unix.S_IFIFO
	default:
		return fmt.Errorf("unsupported node kind %q", string(kind))
	}
	dev := unix.Mkdev(major, minor)
	if err := unix.Mknod(path, modeBits|uint32(mode.Perm()), int(dev)); err != nil {
		if err == unix.EPERM {
			return fmt.Errorf("mknod: operation not permitted (creating %c device nodes needs CAP_MKNOD)", kind)
		}
		return fmt.Errorf("mknod: %w", err)
	}
	return nil
}
