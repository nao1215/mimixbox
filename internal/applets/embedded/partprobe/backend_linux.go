//go:build linux

package partprobe

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// blkRRPart is the BLKRRPART ioctl that re-reads a device's partition table.
const blkRRPart = 0x125f

// osReReader performs the privileged partition-table re-read.
type osReReader struct{}

// ReRead opens the device and issues BLKRRPART.
func (osReReader) ReRead(device string) error {
	f, err := os.OpenFile(device, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied (re-reading a partition table needs privilege)")
		}
		return err
	}
	defer func() { _ = f.Close() }()
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), blkRRPart, 0); errno != 0 {
		if errno == unix.EPERM || errno == unix.EACCES {
			return fmt.Errorf("operation not permitted (re-reading a partition table needs privilege)")
		}
		return errno
	}
	return nil
}
