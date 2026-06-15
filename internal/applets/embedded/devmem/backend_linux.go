//go:build linux

package devmem

import (
	"encoding/binary"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// devMem is the device file mapped for physical access; overridable in tests.
var devMem = "/dev/mem"

// osBackend reads and writes physical memory by mmap-ing /dev/mem. Opening
// /dev/mem requires privilege, so without it the backend returns a documented
// error rather than silently succeeding.
type osBackend struct{}

// Read returns width bytes at the physical address.
func (osBackend) Read(addr uint64, width int) (uint64, error) {
	return access(addr, width, false, 0)
}

// Write stores value (width bytes) at the physical address.
func (osBackend) Write(addr uint64, width int, value uint64) error {
	_, err := access(addr, width, true, value)
	return err
}

// access maps the page containing addr and performs the read or write.
func access(addr uint64, width int, write bool, value uint64) (uint64, error) {
	f, err := os.OpenFile(devMem, os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		if os.IsPermission(err) {
			return 0, fmt.Errorf("open %s: permission denied (physical memory access needs privilege)", devMem)
		}
		return 0, fmt.Errorf("open %s: %w", devMem, err)
	}
	defer func() { _ = f.Close() }()

	pageSize := uint64(unix.Getpagesize())
	base := addr &^ (pageSize - 1)
	off := int(addr - base)

	prot := unix.PROT_READ
	if write {
		prot |= unix.PROT_WRITE
	}
	data, err := unix.Mmap(int(f.Fd()), int64(base), int(pageSize), prot, unix.MAP_SHARED)
	if err != nil {
		return 0, fmt.Errorf("mmap %s: %w", devMem, err)
	}
	defer func() { _ = unix.Munmap(data) }()

	slice := data[off : off+width]
	if write {
		putLE(slice, width, value)
		return 0, nil
	}
	return getLE(slice, width), nil
}

// getLE reads width little-endian bytes as an unsigned integer.
func getLE(b []byte, width int) uint64 {
	switch width {
	case 1:
		return uint64(b[0])
	case 2:
		return uint64(binary.LittleEndian.Uint16(b))
	case 4:
		return uint64(binary.LittleEndian.Uint32(b))
	default:
		return binary.LittleEndian.Uint64(b)
	}
}

// putLE writes value as width little-endian bytes.
func putLE(b []byte, width int, value uint64) {
	switch width {
	case 1:
		b[0] = byte(value)
	case 2:
		binary.LittleEndian.PutUint16(b, uint16(value))
	case 4:
		binary.LittleEndian.PutUint32(b, uint32(value))
	default:
		binary.LittleEndian.PutUint64(b, value)
	}
}
