//go:build linux

package i2c

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// i2cDevPattern formats a bus number into its /dev path; overridable in tests.
var i2cDevPattern = "/dev/i2c-%d"

// I2C ioctl request to bind the open file descriptor to a slave address.
const i2cSlave = 0x0703

// osBackend talks to /dev/i2c-* through the kernel I2C character device. Opening
// the bus and binding a slave address both require privilege; failures surface
// as documented errors instead of silent no-ops.
type osBackend struct{}

// openBus opens /dev/i2c-N and binds it to the slave address.
func openBus(bus, addr int) (*os.File, error) {
	path := fmt.Sprintf(i2cDevPattern, bus)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("open %s: permission denied (I2C access needs privilege)", path)
		}
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("open %s: no such bus", path)
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if addr >= 0 {
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), i2cSlave, uintptr(addr)); errno != 0 {
			_ = f.Close()
			return nil, fmt.Errorf("set I2C slave 0x%02x: %v", addr, errno)
		}
	}
	return f, nil
}

// ReadReg reads one byte from reg (or the current pointer when reg < 0).
func (osBackend) ReadReg(bus, addr, reg int) (byte, error) {
	f, err := openBus(bus, addr)
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	if reg >= 0 {
		if _, err := f.Write([]byte{byte(reg)}); err != nil {
			return 0, fmt.Errorf("select register 0x%02x: %w", reg, err)
		}
	}
	buf := make([]byte, 1)
	if _, err := f.Read(buf); err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}
	return buf[0], nil
}

// WriteReg writes value to reg of the device.
func (osBackend) WriteReg(bus, addr, reg int, value byte) error {
	f, err := openBus(bus, addr)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write([]byte{byte(reg), value}); err != nil {
		return fmt.Errorf("write register 0x%02x: %w", reg, err)
	}
	return nil
}

// Detect probes addresses lo..hi and returns those that respond to a 1-byte
// read.
func (osBackend) Detect(bus, lo, hi int) ([]int, error) {
	// Verify the bus exists/is accessible first so an unreadable bus is a
	// hard error rather than an empty result.
	f, err := openBus(bus, -1)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var found []int
	for addr := lo; addr <= hi; addr++ {
		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), i2cSlave, uintptr(addr)); errno != 0 {
			continue
		}
		buf := make([]byte, 1)
		if _, err := f.Read(buf); err == nil {
			found = append(found, addr)
		}
	}
	return found, nil
}
