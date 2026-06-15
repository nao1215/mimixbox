//go:build linux

package raidautorun

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// raidAutorun is the RAID_AUTORUN md ioctl.
const raidAutorun = 0x934

// osAutoRunner issues RAID_AUTORUN on an md device.
type osAutoRunner struct{}

// Run opens the md device and triggers autodetection.
func (osAutoRunner) Run(device string) error {
	f, err := os.OpenFile(device, os.O_RDONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied (RAID autorun needs privilege)")
		}
		return err
	}
	defer func() { _ = f.Close() }()
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), raidAutorun, 0); errno != 0 {
		if errno == unix.EPERM || errno == unix.EACCES {
			return fmt.Errorf("operation not permitted (RAID autorun needs privilege)")
		}
		return errno
	}
	return nil
}
