//go:build linux

package watchdog

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Watchdog ioctls and the magic close byte from <linux/watchdog.h>.
const (
	wdiocSetTimeout = 0xc0045706
	wdiocKeepAlive  = 0x80045705
	magicClose      = 'V'
)

// osPinger drives a real /dev/watchdog device.
type osPinger struct{}

// Open opens the watchdog and programs its timeout.
func (osPinger) Open(device string, timeoutSec int) (func() error, func() error, error) {
	f, err := os.OpenFile(device, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return nil, nil, fmt.Errorf("permission denied (opening a watchdog needs privilege)")
		}
		return nil, nil, err
	}
	to := int32(timeoutSec)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), wdiocSetTimeout, uintptr(unsafe.Pointer(&to))); errno != 0 {
		_ = f.Close()
		return nil, nil, fmt.Errorf("set timeout: %v", errno)
	}

	keepalive := func() error {
		var arg int32
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), wdiocKeepAlive, uintptr(unsafe.Pointer(&arg))); errno != 0 {
			return errno
		}
		return nil
	}
	closeFn := func() error {
		// Writing the magic 'V' before close requests a graceful stop so the
		// hardware does not reset the system after we exit.
		_, _ = f.Write([]byte{magicClose})
		return f.Close()
	}
	return keepalive, closeFn, nil
}
