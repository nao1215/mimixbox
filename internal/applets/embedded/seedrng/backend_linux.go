//go:build linux

package seedrng

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// rndAddEntropy is the RNDADDENTROPY ioctl that adds buffer contents to the
// kernel entropy pool and bumps the entropy count.
const rndAddEntropy = 0x40085203

// osSeeder feeds the kernel RNG via /dev/urandom and RNDADDENTROPY.
type osSeeder struct{}

// rndEntropy mirrors struct rand_pool_info for the RNDADDENTROPY ioctl.
type rndEntropy struct {
	entropyCount int32
	bufSize      int32
	buf          [256]byte
}

// Credit feeds seed into the kernel pool.
func (osSeeder) Credit(seed []byte, credit bool) error {
	f, err := os.OpenFile("/dev/urandom", os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied (seeding the kernel RNG needs privilege)")
		}
		return err
	}
	defer func() { _ = f.Close() }()

	var info rndEntropy
	n := copy(info.buf[:], seed)
	info.bufSize = int32(n)
	if credit {
		info.entropyCount = int32(n * 8)
	}
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), rndAddEntropy, uintptr(unsafe.Pointer(&info))); errno != 0 {
		if errno == unix.EPERM {
			return fmt.Errorf("operation not permitted (crediting kernel entropy needs CAP_SYS_ADMIN)")
		}
		return errno
	}
	return nil
}
