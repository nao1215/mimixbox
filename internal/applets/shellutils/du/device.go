package du

import (
	"fmt"
	"os"
	"syscall"
)

// deviceOf returns the filesystem device id (st_dev) for path. It is used by
// --one-file-system to detect when recursion would cross a mount point. The
// du package targets Linux (consistent with the rest of MimixBox), so a
// syscall.Stat_t assertion needs no build tags.
func deviceOf(path string) (uint64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("cannot determine device of %s", path)
	}
	return uint64(st.Dev), nil
}
