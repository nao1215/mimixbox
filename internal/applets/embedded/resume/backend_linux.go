//go:build linux

package resume

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// osResolver resolves a block device's major:minor from its stat data.
type osResolver struct{}

// DevNumber stats device and returns its "major:minor".
func (osResolver) DevNumber(device string) (string, error) {
	fi, err := os.Stat(device)
	if err != nil {
		return "", err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("cannot read device numbers of %s", device)
	}
	if fi.Mode()&os.ModeDevice == 0 {
		return "", fmt.Errorf("%s is not a block device", device)
	}
	return fmt.Sprintf("%d:%d", unix.Major(uint64(st.Rdev)), unix.Minor(uint64(st.Rdev))), nil
}
